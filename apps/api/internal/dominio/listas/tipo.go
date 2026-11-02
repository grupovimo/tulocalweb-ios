package listas

import (
	"errors"
	"strconv"
	"strings"
)

// Tipos de producto (migración 00071). El editor pregunta primero qué es y según eso pide sus datos:
// la ropa, talle y color; la comida, el contenido; un mueble, peso y medidas. La misma lista está en
// apps/web/src/lib/tienda/tipo-producto.ts y en la app de iPhone (TipoProducto.swift).
var tiposProducto = map[string]bool{
	"ropa": true, "calzado": true, "alimentos": true, "belleza": true, "hogar": true, "tecnologia": true, "otro": true,
}

// atributosProducto: los datos que se pueden cargar; los numéricos van como número en texto.
var atributosProducto = map[string]bool{
	"peso": true, "alto": true, "ancho": true, "largo": true, // gramos y centímetros
	"contenido": true, "material": true, "marca": true, "modelo": true, "garantia": true,
}

var numericos = map[string]bool{"peso": true, "alto": true, "ancho": true, "largo": true}

// limpiarTipo revisa el tipo y los atributos: claves conocidas, textos de hasta 60 letras, peso y
// medidas como enteros positivos. Los vacíos se descartan.
func limpiarTipo(tipo *string, atributos map[string]string) (*string, map[string]string, error) {
	if tipo != nil {
		t := strings.TrimSpace(*tipo)
		if t == "" {
			tipo = nil
		} else if !tiposProducto[t] {
			return nil, nil, errors.New("el tipo de producto no existe")
		} else {
			tipo = &t
		}
	}
	if atributos == nil {
		return tipo, nil, nil
	}
	limpios := map[string]string{}
	for k, v := range atributos {
		if !atributosProducto[k] {
			return nil, nil, errors.New("hay un dato del producto que no existe: " + k)
		}
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if len([]rune(v)) > 60 {
			return nil, nil, errors.New("cada dato del producto puede tener hasta 60 letras")
		}
		if numericos[k] {
			n, err := strconv.Atoi(v)
			if err != nil || n <= 0 || n > 10_000_000 {
				return nil, nil, errors.New("el peso y las medidas van en números enteros (gramos y centímetros)")
			}
			v = strconv.Itoa(n)
		}
		limpios[k] = v
	}
	return tipo, limpios, nil
}
