// Package correo manda mails transaccionales (recuperar contraseña, avisos). Tres implementaciones:
// Resend, SMTP (Mailpit en desarrollo, o el servidor de correo propio con usuario y TLS) y Memoria
// en los tests.
package correo

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mime"
	"net"
	"net/http"
	"net/smtp"
	"sort"
	"strings"
	"sync"
	"time"
)

// Mensaje es un mail. HTML y Texto van los dos (los clientes que no muestran HTML usan el texto).
// ResponderA (opcional): a dónde va la respuesta, p. ej. la tienda en los mails a quien le compró.
// DeNombre (opcional) reemplaza el nombre del remitente (la dirección sigue siendo la de la
// plataforma): el newsletter sale con el nombre de la tienda. Encabezados suma encabezados propios
// (List-Unsubscribe del newsletter).
type Mensaje struct {
	Para, Asunto, HTML, Texto string
	ResponderA                string
	DeNombre                  string
	Encabezados               map[string]string
}

// remitente arma el From con el nombre del mensaje, si trae uno.
func remitente(de string, m Mensaje) string {
	nombre := strings.Map(func(r rune) rune {
		// Sin los especiales de RFC 5322: el nombre va sin comillas.
		if strings.ContainsRune("\"<>\\,;:()[]@\r\n", r) {
			return -1
		}
		return r
	}, strings.TrimSpace(m.DeNombre))
	if nombre == "" {
		return de
	}
	return nombre + " <" + direccionDe(de) + ">"
}

// Correo manda un mensaje.
type Correo interface {
	Enviar(ctx context.Context, m Mensaje) error
}

// Resend usa la API HTTP de Resend.
type Resend struct {
	Clave, De string
	HTTP      *http.Client
	URL       string // por defecto https://api.resend.com/emails
}

func (r *Resend) Enviar(ctx context.Context, m Mensaje) error {
	url := r.URL
	if url == "" {
		url = "https://api.resend.com/emails"
	}
	datos := map[string]any{"from": remitente(r.De, m), "to": []string{m.Para}, "subject": m.Asunto, "html": m.HTML, "text": m.Texto}
	if m.ResponderA != "" {
		datos["reply_to"] = m.ResponderA
	}
	if len(m.Encabezados) > 0 {
		datos["headers"] = m.Encabezados
	}
	cuerpo, _ := json.Marshal(datos)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(cuerpo))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+r.Clave)
	req.Header.Set("Content-Type", "application/json")
	c := r.HTTP
	if c == nil {
		c = &http.Client{Timeout: 10 * time.Second}
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		var b bytes.Buffer
		_, _ = b.ReadFrom(res.Body)
		return fmt.Errorf("resend respondió %d: %s", res.StatusCode, strings.TrimSpace(b.String()))
	}
	return nil
}

// SMTP manda por un servidor SMTP. Sin Usuario, sin autenticación: Mailpit en desarrollo
// (infra/compose.dev.yml). Con Usuario y Clave, el servidor de correo propio: en el puerto 465 con
// TLS desde el arranque (submissions) y en los demás con STARTTLS.
type SMTP struct{ Direccion, De, Usuario, Clave string }

// Elegir arma el correo según la configuración: Resend si hay clave; si no, SMTP.
func Elegir(resendClave, smtpDireccion, usuario, clave, de string) Correo {
	if resendClave != "" {
		return &Resend{Clave: resendClave, De: de}
	}
	return &SMTP{Direccion: smtpDireccion, De: de, Usuario: usuario, Clave: clave}
}

func (s *SMTP) Enviar(_ context.Context, m Mensaje) error {
	// Para y Asunto van en los encabezados: un salto de línea metería encabezados ajenos.
	if strings.ContainsAny(m.Para+m.Asunto+m.ResponderA, "\r\n") {
		return fmt.Errorf("correo: salto de línea en Para, Asunto o ResponderA")
	}
	for k, v := range m.Encabezados {
		if strings.ContainsAny(k+v, "\r\n") || strings.ContainsAny(k, ": ") {
			return fmt.Errorf("correo: encabezado inválido")
		}
	}
	frontera := "mostrador-frontera"
	var b strings.Builder
	// El asunto va codificado (RFC 2047): con tildes en crudo, varios clientes lo muestran roto.
	fmt.Fprintf(&b, "From: %s\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nMessage-ID: %s\r\nMIME-Version: 1.0\r\n",
		de(remitente(s.De, m)), m.Para, mime.QEncoding.Encode("utf-8", m.Asunto), time.Now().Format(time.RFC1123Z), idMensaje(s.De))
	if m.ResponderA != "" {
		fmt.Fprintf(&b, "Reply-To: %s\r\n", m.ResponderA)
	}
	claves := make([]string, 0, len(m.Encabezados))
	for k := range m.Encabezados {
		claves = append(claves, k)
	}
	sort.Strings(claves)
	for _, k := range claves {
		fmt.Fprintf(&b, "%s: %s\r\n", k, m.Encabezados[k])
	}
	fmt.Fprintf(&b, "Content-Type: multipart/alternative; boundary=%s\r\n\r\n", frontera)
	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s\r\n", frontera, m.Texto)
	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/html; charset=utf-8\r\n\r\n%s\r\n--%s--\r\n", frontera, m.HTML, frontera)
	if s.Usuario == "" {
		return smtp.SendMail(s.Direccion, nil, direccionDe(s.De), []string{m.Para}, []byte(b.String()))
	}
	return s.conUsuario(m.Para, []byte(b.String()))
}

// conUsuario manda autenticado y siempre cifrado: nunca manda la clave en claro.
func (s *SMTP) conUsuario(para string, cuerpo []byte) error {
	host, puerto, err := net.SplitHostPort(s.Direccion)
	if err != nil {
		return fmt.Errorf("correo: dirección SMTP inválida: %w", err)
	}
	cfg := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
	d := &net.Dialer{Timeout: 20 * time.Second}
	var c *smtp.Client
	if puerto == "465" {
		con, err := tls.DialWithDialer(d, "tcp", s.Direccion, cfg)
		if err != nil {
			return err
		}
		if c, err = smtp.NewClient(con, host); err != nil {
			return err
		}
	} else {
		con, err := d.Dial("tcp", s.Direccion)
		if err != nil {
			return err
		}
		if c, err = smtp.NewClient(con, host); err != nil {
			return err
		}
		if err := c.StartTLS(cfg); err != nil {
			return fmt.Errorf("correo: el servidor no acepta STARTTLS: %w", err)
		}
	}
	defer c.Close()
	if err := c.Auth(smtp.PlainAuth("", s.Usuario, s.Clave, host)); err != nil {
		return err
	}
	if err := c.Mail(direccionDe(s.De)); err != nil {
		return err
	}
	if err := c.Rcpt(para); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(cuerpo); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

// de codifica el nombre del remitente si tiene caracteres fuera de ASCII ("Tu Local Web <hola@…>").
func de(remitente string) string {
	i := strings.LastIndex(remitente, "<")
	if i <= 0 {
		return remitente
	}
	return mime.QEncoding.Encode("utf-8", strings.TrimSpace(remitente[:i])) + " " + remitente[i:]
}

// idMensaje arma un Message-ID único con el dominio del remitente (sin él, los filtros de spam restan).
func idMensaje(remitente string) string {
	r := make([]byte, 12)
	_, _ = rand.Read(r)
	dominio := "localhost"
	if d := direccionDe(remitente); strings.Contains(d, "@") {
		dominio = d[strings.LastIndex(d, "@")+1:]
	}
	return "<" + hex.EncodeToString(r) + "@" + dominio + ">"
}

func direccionDe(de string) string {
	if i := strings.LastIndex(de, "<"); i >= 0 {
		return strings.TrimSuffix(de[i+1:], ">")
	}
	return de
}

// Memoria guarda los mensajes (tests).
type Memoria struct {
	mu       sync.Mutex
	Enviados []Mensaje
}

func (m *Memoria) Enviar(_ context.Context, msj Mensaje) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Enviados = append(m.Enviados, msj)
	return nil
}

// Ultimo devuelve el último mensaje mandado a esa dirección.
func (m *Memoria) Ultimo(para string) (Mensaje, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := len(m.Enviados) - 1; i >= 0; i-- {
		if m.Enviados[i].Para == para {
			return m.Enviados[i], true
		}
	}
	return Mensaje{}, false
}
