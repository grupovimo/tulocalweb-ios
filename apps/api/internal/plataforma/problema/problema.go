// Package problema escribe los errores de la API como application/problem+json (RFC 9457), en
// criollo y con un código estable que el front puede mapear.
package problema

import (
	"encoding/json"
	"net/http"
)

// Problema es un error para el cliente. Detalle se muestra tal cual: tiene que servirle a una persona.
type Problema struct {
	Tipo    string  `json:"type"`
	Titulo  string  `json:"title"`
	Estado  int     `json:"status"`
	Detalle string  `json:"detail,omitempty"`
	Codigo  string  `json:"code,omitempty"`
	Campos  []Campo `json:"issues,omitempty"`
	// Restantes: intentos que quedan antes del bloqueo (sólo en el ingreso con contraseña).
	Restantes *int `json:"restantes,omitempty"`
}

// Campo es un error de validación de un campo puntual.
type Campo struct {
	Ruta    string `json:"path"`
	Mensaje string `json:"message"`
}

func (p *Problema) Error() string { return p.Titulo + ": " + p.Detalle }

// Escribir manda el problema como respuesta.
func Escribir(w http.ResponseWriter, p *Problema) {
	if p.Tipo == "" {
		p.Tipo = "about:blank"
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(p.Estado)
	_ = json.NewEncoder(w).Encode(p)
}

func nuevo(estado int, codigo, titulo, detalle string) *Problema {
	return &Problema{Estado: estado, Codigo: codigo, Titulo: titulo, Detalle: detalle}
}

// Constructores de los casos comunes.
func Invalido(detalle string, campos ...Campo) *Problema {
	p := nuevo(http.StatusBadRequest, "invalido", "Revisá los datos", detalle)
	p.Campos = campos
	return p
}
func NoEncontrado(detalle string) *Problema {
	return nuevo(http.StatusNotFound, "no_encontrado", "No lo encontramos", detalle)
}
func SinSesion() *Problema {
	return nuevo(http.StatusUnauthorized, "sin_sesion", "Tenés que ingresar", "Tu sesión venció o no ingresaste.")
}
func SinPermiso(detalle string) *Problema {
	return nuevo(http.StatusForbidden, "sin_permiso", "No tenés permiso", detalle)
}
func Conflicto(codigo, detalle string) *Problema {
	return nuevo(http.StatusConflict, codigo, "No se pudo completar", detalle)
}
func NoProcesable(codigo, detalle string) *Problema {
	return nuevo(http.StatusUnprocessableEntity, codigo, "No se pudo completar", detalle)
}
func Interno() *Problema {
	return nuevo(http.StatusInternalServerError, "interno", "Algo salió mal de nuestro lado",
		"Ya quedó registrado. Probá de nuevo en un rato.")
}
