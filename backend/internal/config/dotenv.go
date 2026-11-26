// Package config carga la configuración local desde un archivo .env.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadDotEnv lee variables desde el primer archivo .env que encuentre entre
// las rutas indicadas. No reemplaza variables que ya estén definidas en el
// sistema, así en un servidor mandan las variables reales del entorno.
// Si no existe ningún archivo .env, no hace nada.
func LoadDotEnv(paths ...string) (string, error) {
	for _, path := range paths {
		file, err := os.Open(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNumber := 0
		for scanner.Scan() {
			lineNumber++
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			line = strings.TrimPrefix(line, "export ")
			key, value, found := strings.Cut(line, "=")
			if !found {
				return "", fmt.Errorf("%s línea %d: formato inválido, se esperaba CLAVE=valor", path, lineNumber)
			}
			key = strings.TrimSpace(key)
			value = cleanValue(value)
			if _, exists := os.LookupEnv(key); !exists {
				os.Setenv(key, value)
			}
		}
		return path, scanner.Err()
	}
	return "", nil
}

// cleanValue quita comillas y comentarios al final de la línea.
func cleanValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
		return value[1 : len(value)-1]
	}
	if i := strings.Index(value, " #"); i >= 0 {
		value = value[:i]
	}
	return strings.TrimSpace(value)
}
