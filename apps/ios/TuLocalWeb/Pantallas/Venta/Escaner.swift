import AVFoundation
import SwiftUI
import UIKit

/// La cámara para escanear códigos de barras (EAN, UPC, Code 128, QR). Llama a `alLeer` con cada
/// código nuevo (el mismo código no se repite hasta pasado un segundo y medio).
struct Escaner: UIViewRepresentable {
    var linterna: Bool
    var alLeer: (String) -> Void

    func makeCoordinator() -> Coordinador { Coordinador(alLeer: alLeer) }

    func makeUIView(context: Context) -> VistaCamara {
        let v = VistaCamara()
        v.capa.videoGravity = .resizeAspectFill
        context.coordinator.arrancar(en: v)
        return v
    }

    func updateUIView(_ v: VistaCamara, context: Context) {
        context.coordinator.alLeer = alLeer
        context.coordinator.linterna(linterna)
    }

    static func dismantleUIView(_ v: VistaCamara, coordinator: Coordinador) {
        coordinator.parar()
    }

    final class VistaCamara: UIView {
        override class var layerClass: AnyClass { AVCaptureVideoPreviewLayer.self }
        var capa: AVCaptureVideoPreviewLayer { layer as! AVCaptureVideoPreviewLayer }
    }

    final class Coordinador: NSObject, AVCaptureMetadataOutputObjectsDelegate {
        var alLeer: (String) -> Void
        private let sesion = AVCaptureSession()
        private let cola = DispatchQueue(label: "ar.com.tulocalweb.escaner")
        private var ultimo: (String, Date)?

        init(alLeer: @escaping (String) -> Void) { self.alLeer = alLeer }

        func arrancar(en v: VistaCamara) {
            v.capa.session = sesion
            cola.async { [self] in
                guard let disp = AVCaptureDevice.default(for: .video),
                      let entrada = try? AVCaptureDeviceInput(device: disp), sesion.canAddInput(entrada) else { return }
                sesion.beginConfiguration()
                sesion.addInput(entrada)
                let salida = AVCaptureMetadataOutput()
                if sesion.canAddOutput(salida) {
                    sesion.addOutput(salida)
                    salida.setMetadataObjectsDelegate(self, queue: .main)
                    let tipos: [AVMetadataObject.ObjectType] = [.ean13, .ean8, .upce, .code128, .code39, .qr]
                    salida.metadataObjectTypes = tipos.filter { salida.availableMetadataObjectTypes.contains($0) }
                }
                sesion.commitConfiguration()
                sesion.startRunning()
            }
        }

        func parar() {
            cola.async { [sesion] in sesion.stopRunning() }
        }

        func linterna(_ prendida: Bool) {
            guard let d = AVCaptureDevice.default(for: .video), d.hasTorch, (d.torchMode == .on) != prendida else { return }
            try? d.lockForConfiguration()
            d.torchMode = prendida ? .on : .off
            d.unlockForConfiguration()
        }

        func metadataOutput(_ output: AVCaptureMetadataOutput, didOutput objetos: [AVMetadataObject], from connection: AVCaptureConnection) {
            guard let o = objetos.first as? AVMetadataMachineReadableCodeObject, let codigo = o.stringValue else { return }
            if let u = ultimo, u.0 == codigo, Date().timeIntervalSince(u.1) < 1.5 { return }
            ultimo = (codigo, Date())
            alLeer(codigo)
        }
    }
}
