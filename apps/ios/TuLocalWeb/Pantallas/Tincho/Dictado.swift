import AVFoundation
import Speech
import SwiftUI

/// 6a · El micrófono del chat de Tincho: dictado en español (Argentina) que va escribiendo en la
/// entrada mientras se habla. Pide los dos permisos la primera vez; se corta solo con 2 s de
/// silencio o al tocar de nuevo.
@Observable
final class Dictado {
    var escuchando = false
    var texto = ""
    var error: String?

    private let reconocedor = SFSpeechRecognizer(locale: Locale(identifier: "es-AR"))
    private let motor = AVAudioEngine()
    private var pedido: SFSpeechAudioBufferRecognitionRequest?
    private var tarea: SFSpeechRecognitionTask?
    private var silencio: Task<Void, Never>?

    func alternar() async {
        if escuchando { parar(); return }
        error = nil
        guard await permisos() else {
            error = "Para dictar, habilitá el micrófono y el reconocimiento de voz en Ajustes."
            return
        }
        do { try empezar() } catch {
            parar()
            self.error = "No pudimos usar el micrófono ahora."
        }
    }

    private func permisos() async -> Bool {
        let voz = await withCheckedContinuation { c in
            SFSpeechRecognizer.requestAuthorization { c.resume(returning: $0 == .authorized) }
        }
        guard voz else { return false }
        return await AVAudioApplication.requestRecordPermission()
    }

    private func empezar() throws {
        guard let reconocedor, reconocedor.isAvailable else { throw CancellationError() }
        let sesion = AVAudioSession.sharedInstance()
        try sesion.setCategory(.record, mode: .measurement, options: .duckOthers)
        try sesion.setActive(true, options: .notifyOthersOnDeactivation)
        let p = SFSpeechAudioBufferRecognitionRequest()
        p.shouldReportPartialResults = true
        if reconocedor.supportsOnDeviceRecognition { p.requiresOnDeviceRecognition = true }
        pedido = p
        let entrada = motor.inputNode
        entrada.removeTap(onBus: 0)
        entrada.installTap(onBus: 0, bufferSize: 1024, format: entrada.outputFormat(forBus: 0)) { [weak p] b, _ in p?.append(b) }
        motor.prepare()
        try motor.start()
        texto = ""
        escuchando = true
        tarea = reconocedor.recognitionTask(with: p) { [weak self] r, e in
            DispatchQueue.main.async {
                guard let self else { return }
                if let r {
                    self.texto = r.bestTranscription.formattedString
                    self.reiniciarSilencio()
                    if r.isFinal { self.parar() }
                }
                if e != nil { self.parar() }
            }
        }
        reiniciarSilencio()
    }

    private func reiniciarSilencio() {
        silencio?.cancel()
        silencio = Task { @MainActor [weak self] in
            try? await Task.sleep(for: .seconds(2))
            guard !Task.isCancelled else { return }
            self?.parar()
        }
    }

    func parar() {
        silencio?.cancel()
        if motor.isRunning {
            motor.stop()
            motor.inputNode.removeTap(onBus: 0)
        }
        pedido?.endAudio()
        tarea?.cancel()
        pedido = nil
        tarea = nil
        escuchando = false
        try? AVAudioSession.sharedInstance().setActive(false, options: .notifyOthersOnDeactivation)
    }
}
