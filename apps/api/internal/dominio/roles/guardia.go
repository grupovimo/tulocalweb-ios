package roles

import (
	"net/http"
	"strings"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Rutas: el permiso de cada ruta del panel, con el mismo patrón con que se registra en el mux.
// Un test recorre el código y falla si una ruta de /v1/panel/ no está acá. Las que dependen de un
// valor del camino (la lista, el grupo de opciones) tienen además su patrón más específico.
var Rutas = map[string]Permiso{
	// Lo de cualquiera del equipo.
	"GET /v1/panel/tienda":                          Panel,
	"GET /v1/panel/resumen":                         Panel,
	"GET /v1/panel/tablero":                         Panel,
	"POST /v1/panel/avisos/{aviso}/ocultar":         Panel,
	"GET /v1/panel/notificaciones":                  Panel, // las de cada persona (§45)
	"GET /v1/panel/notificaciones/sin-leer":         Panel,
	"POST /v1/panel/notificaciones/{id}/leida":      Panel,
	"POST /v1/panel/notificaciones/leidas":          Panel,
	"GET /v1/panel/para-hacer":                      Panel,
	"PUT /v1/panel/para-hacer/{id}":                 Panel,
	"GET /v1/panel/hornero/contexto":                Panel,
	"POST /v1/panel/hornero/ideas/{idea}":           Panel,
	"GET /v1/panel/configuracion":                   Panel, // sin negocio ni equipo para quien no puede (configuracion.Leer)
	"PUT /v1/panel/configuracion/seguridad/{clave}": Panel, // la seguridad es de la persona
	"DELETE /v1/panel/sesiones":                     Panel,
	"DELETE /v1/panel/sesiones/{id}":                Panel,
	"GET /v1/panel/turnos/ajustes":                  Panel, // el armazón decide si muestra Turnos
	"POST /v1/panel/dispositivos":                   Panel, // el iPhone de la sesión (push)
	"DELETE /v1/panel/dispositivos":                 Panel,
	"GET /v1/panel/push-web":                        Panel, // el panel instalado de la sesión (push web)
	"POST /v1/panel/push-web":                       Panel,
	"DELETE /v1/panel/push-web":                     Panel,
	"POST /v1/panel/push-web/prueba":                Panel,

	// Pedidos, punto de venta, envíos, devoluciones y turnos.
	"GET /v1/panel/pedidos":                                Pedidos,
	"POST /v1/panel/pedidos":                               Pedidos,
	"POST /v1/panel/pedidos/lote":                          Pedidos,
	"GET /v1/panel/pedidos/{numero}":                       Pedidos,
	"GET /v1/panel/pedidos/{numero}/correo":                Pedidos,
	"POST /v1/panel/pedidos/{numero}/correo":               Pedidos,
	"PUT /v1/panel/pedidos/{numero}/correo/seguimiento":    Pedidos,
	"POST /v1/panel/pedidos/{numero}/estado":               Pedidos,
	"POST /v1/panel/pedidos/{numero}/saldado":              Pedidos,
	"DELETE /v1/panel/pedidos/{numero}/saldado":            Pedidos,
	"POST /v1/panel/pedidos/{numero}/qr":                   Pedidos,
	"POST /v1/panel/pedidos/{numero}/recordatorio":         Pedidos,
	"GET /v1/panel/cobros/qr/{referencia}":                 Pedidos,
	"DELETE /v1/panel/cobros/qr/{referencia}":              Pedidos,
	"POST /v1/panel/links":                                 Pedidos,
	"GET /v1/panel/devoluciones":                           Pedidos,
	"POST /v1/panel/devoluciones/{id}":                     Pedidos,
	"GET /v1/panel/turnos":                                 Pedidos,
	"POST /v1/panel/turnos":                                Pedidos,
	"PATCH /v1/panel/turnos/{id}":                          Pedidos,
	"GET /v1/panel/turnos/profesionales":                   Pedidos,
	"GET /v1/panel/turnos/servicios":                       Pedidos,
	"GET /v1/panel/listas/{lista}":                         Pedidos, // productos, inventario y envíos
	"POST /v1/panel/listas/{lista}/acciones":               Pedidos,
	"POST /v1/panel/listas/envios/acciones":                Pedidos,
	"GET /v1/panel/inventario/movimientos":                 Pedidos,
	"GET /v1/panel/productos/{id}":                         Pedidos, // sin costo para quien no tiene Catalogo
	"GET /v1/panel/productos/{id}/lanzamiento":             Pedidos,
	"GET /v1/panel/configuracion/entregas":                 Pedidos,
	"GET /v1/panel/tarjeta":                                Pedidos,
	"GET /v1/panel/listas/clientes":                        Clientes,
	"POST /v1/panel/listas/clientes/acciones":              Clientes,
	"GET /v1/panel/clientes/{id}/historial":                Clientes,
	"GET /v1/panel/bandeja":                                Clientes,
	"GET /v1/panel/bandeja/{id}":                           Clientes,
	"POST /v1/panel/bandeja/{id}":                          Clientes,
	"GET /v1/panel/carritos":                               Clientes,
	"POST /v1/panel/carritos/recordar":                     Clientes,
	"POST /v1/panel/carritos/{id}/recordar":                Clientes,
	"GET /v1/panel/resenas":                                Clientes,
	"POST /v1/panel/resenas/{id}":                          Clientes,
	"POST /v1/panel/preguntas/{id}":                        Clientes,
	"GET /v1/panel/mercadolibre/preguntas":                 Clientes,
	"POST /v1/panel/mercadolibre/preguntas/{id}/respuesta": Clientes,

	// Productos y stock (los precios y costos los controla listas.Guardar).
	"PUT /v1/panel/productos/{id}":                  Productos,
	"POST /v1/panel/listas/productos/acciones":      Productos, // «precio» pide Catalogo (listas.Rutas)
	"POST /v1/panel/listas/inventario/acciones":     Productos,
	"POST /v1/panel/productos":                      Catalogo,
	"PUT /v1/panel/productos/{id}/lanzamiento":      Catalogo, // precio de oferta y señas
	"DELETE /v1/panel/productos/{id}":               Catalogo,
	"POST /v1/panel/productos/{id}/duplicar":        Catalogo,
	"POST /v1/panel/productos/{id}/mercadolibre":    Catalogo,
	"DELETE /v1/panel/productos/{id}/mercadolibre":  Catalogo,
	"GET /v1/panel/listas-precio":                   Catalogo,
	"POST /v1/panel/listas-precio/aumento":          Catalogo,
	"POST /v1/panel/listas-precio/aumento/deshacer": Catalogo,
	"DELETE /v1/panel/listas-precio/programado":     Catalogo,
	"POST /v1/panel/relampago":                      Catalogo,
	"PUT /v1/panel/automaticas/{id}":                Catalogo,
	"GET /v1/panel/cupones":                         Catalogo,
	"POST /v1/panel/cupones":                        Catalogo,
	"DELETE /v1/panel/cupones/consejo":              Catalogo,
	"POST /v1/panel/cupones/{id}":                   Catalogo,

	// Plata.
	"GET /v1/panel/finanzas":                            Finanzas,
	"POST /v1/panel/finanzas/retiro":                    Finanzas,
	"GET /v1/panel/informes":                            Finanzas,
	"PUT /v1/panel/meta":                                Finanzas,
	"GET /v1/panel/arca":                                Finanzas,
	"PUT /v1/panel/arca/ajustes":                        Finanzas,
	"PUT /v1/panel/arca/certificado":                    Finanzas,
	"POST /v1/panel/arca/pedido":                        Finanzas,
	"POST /v1/panel/arca/probar":                        Finanzas,
	"GET /v1/panel/listas/facturacion":                  Finanzas,
	"POST /v1/panel/listas/facturacion/acciones":        Finanzas,
	"POST /v1/panel/pedidos/{numero}/reembolso":         Finanzas,
	"POST /v1/panel/turnos/{id}/devolver-sena":          Finanzas,
	"GET /v1/panel/integraciones/mercadopago":           Finanzas,
	"DELETE /v1/panel/integraciones/mercadopago":        Finanzas,
	"POST /v1/panel/integraciones/mercadopago/caja":     Finanzas,
	"POST /v1/panel/integraciones/mercadopago/conectar": Finanzas,
	"POST /v1/panel/integraciones/mercadopago/vuelta":   Finanzas,

	// Configuración de la tienda.
	"PUT /v1/panel/configuracion/negocio":                  Tienda, // incluye alias y CBU
	"PUT /v1/panel/configuracion/{grupo}/{clave}":          Tienda, // avisos de la tienda
	"PUT /v1/panel/configuracion/entregas":                 Tienda,
	"GET /v1/panel/onboarding":                             Tienda,
	"PUT /v1/panel/onboarding":                             Tienda,
	"PUT /v1/panel/tienda/origen":                          Tienda,
	"GET /v1/panel/tienda/direccion":                       Tienda,
	"PUT /v1/panel/tienda/direccion":                       Tienda,
	"GET /v1/panel/actividad":                              Tienda, // quién hizo qué: dueña y administración
	"GET /v1/panel/actividad/exportar":                     Tienda,
	"POST /v1/panel/actividad/{id}/deshacer":               Tienda,
	"GET /v1/panel/datos":                                  Tienda,
	"GET /v1/panel/datos/{tipo}":                           Tienda,
	"POST /v1/panel/datos/exportaciones":                   Tienda,
	"POST /v1/panel/datos/inactivos":                       Tienda,
	"DELETE /v1/panel/datos/inactivos":                     Tienda,
	"GET /v1/panel/dominios":                               Tienda,
	"POST /v1/panel/dominios":                              Tienda,
	"PUT /v1/panel/dominios/{dominio}":                     Tienda,
	"DELETE /v1/panel/dominios/{dominio}":                  Tienda,
	"GET /v1/panel/canales":                                Tienda,
	"PUT /v1/panel/canales/{id}":                           Tienda,
	"GET /v1/panel/marketing":                              Tienda,
	"POST /v1/panel/campanas":                              Tienda,
	"PUT /v1/panel/carritos":                               Tienda,
	"GET /v1/panel/mi-tienda":                              Tienda,
	"PUT /v1/panel/mi-tienda":                              Tienda,
	"POST /v1/panel/mi-tienda/publicar":                    Tienda,
	"PUT /v1/panel/tarjeta":                                Tienda,
	"GET /v1/panel/notificaciones/preferencias":            Tienda,
	"PUT /v1/panel/notificaciones/preferencias":            Tienda,
	"GET /v1/panel/ia":                                     Tienda,
	"DELETE /v1/panel/ia/conexiones/{id}":                  Tienda,
	"PUT /v1/panel/ia/permisos/{permiso}":                  Tienda,
	"GET /v1/panel/telegram":                               Tienda,
	"DELETE /v1/panel/telegram":                            Tienda,
	"POST /v1/panel/telegram/codigo":                       Tienda,
	"POST /v1/panel/importaciones":                         Tienda,
	"GET /v1/panel/importaciones/actual":                   Tienda,
	"GET /v1/panel/importaciones/tiendanube":               Tienda,
	"POST /v1/panel/importaciones/tiendanube/conectar":     Tienda,
	"POST /v1/panel/importaciones/tiendanube/vuelta":       Tienda,
	"POST /v1/panel/importaciones/{id}/cancelar":           Tienda,
	"GET /v1/panel/integraciones/correo":                   Tienda,
	"POST /v1/panel/integraciones/correo":                  Tienda,
	"PUT /v1/panel/integraciones/correo":                   Tienda,
	"DELETE /v1/panel/integraciones/correo":                Tienda,
	"GET /v1/panel/integraciones/medicion":                 Tienda,
	"PUT /v1/panel/integraciones/medicion":                 Tienda,
	"GET /v1/panel/integraciones/medicion/conversiones":    Tienda,
	"PUT /v1/panel/integraciones/medicion/conversiones":    Tienda,
	"DELETE /v1/panel/integraciones/medicion/conversiones": Tienda,
	"GET /v1/panel/integraciones/mercadolibre":             Tienda,
	"PUT /v1/panel/integraciones/mercadolibre":             Tienda,
	"DELETE /v1/panel/integraciones/mercadolibre":          Tienda,
	"POST /v1/panel/integraciones/mercadolibre/conectar":   Tienda,
	"POST /v1/panel/integraciones/mercadolibre/vuelta":     Tienda,
	"PUT /v1/panel/turnos/ajustes":                         Tienda,
	"POST /v1/panel/turnos/profesionales":                  Tienda,
	"PUT /v1/panel/turnos/profesionales/{id}":              Tienda,
	"DELETE /v1/panel/turnos/profesionales/{id}":           Tienda,
	"POST /v1/panel/turnos/servicios":                      Tienda,
	"PUT /v1/panel/turnos/servicios/{id}":                  Tienda,
	"DELETE /v1/panel/turnos/servicios/{id}":               Tienda,

	// Sólo la dueña.
	"GET /v1/panel/cuenta/plan":           Cuenta,
	"PUT /v1/panel/plan":                  Cuenta,
	"POST /v1/panel/cuenta/pack":          Cuenta,
	"DELETE /v1/panel/cuenta/pack":        Cuenta,
	"POST /v1/panel/cuenta/cancelacion":   Cuenta,
	"DELETE /v1/panel/cuenta/cancelacion": Cuenta,
	"POST /v1/panel/cuenta/suscripcion":   Cuenta,
	"DELETE /v1/panel/cuenta/suscripcion": Cuenta,
	"DELETE /v1/panel/cuenta":             Cuenta,
	"POST /v1/panel/equipo":               Cuenta,
	"PUT /v1/panel/equipo/{id}":           Cuenta,
	"DELETE /v1/panel/equipo/{id}":        Cuenta,
	"POST /v1/panel/equipo/{id}/reenviar": Cuenta,
}

// marca es el handler del mux interno: sólo lleva el permiso.
type marca Permiso

func (marca) ServeHTTP(http.ResponseWriter, *http.Request) {}

// Guardia aplica el rol a cada pedido de /v1/panel/. rol devuelve el rol de la sesión y si hay
// una tienda (sin sesión o sin tienda sigue de largo: el handler responde 401). La ruta se resuelve
// con un ServeMux propio, con las mismas reglas que el de la API; una ruta que no está en la tabla
// sólo la puede usar la dueña (cerrado por defecto).
func Guardia(rol func(*http.Request) (string, bool), next http.Handler) http.Handler {
	mux := http.NewServeMux()
	for patron, p := range Rutas {
		mux.Handle(patron, marca(p))
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/panel/") {
			next.ServeHTTP(w, r)
			return
		}
		quien, hay := rol(r)
		if !hay {
			next.ServeHTTP(w, r)
			return
		}
		p := Cuenta
		if h, _ := mux.Handler(r); h != nil {
			if m, ok := h.(marca); ok {
				p = Permiso(m)
			}
		}
		if !Puede(quien, p) {
			problema.Escribir(w, problema.SinPermiso(Mensaje(p)))
			return
		}
		next.ServeHTTP(w, r)
	})
}
