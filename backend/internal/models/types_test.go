package models

import (
	"encoding/json"
	"testing"
)

func TestStringSlice_ScanAndValue(t *testing.T) {
	// Test Value con nil
	var s StringSlice
	val, err := s.Value()
	if err != nil {
		t.Fatalf("Error inesperado en Value con nil: %v", err)
	}
	if val != "[]" {
		t.Fatalf("Esperado '[]', obtenido '%v'", val)
	}

	// Test MarshalJSON con nil
	jsonBytes, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Error marshaling nil StringSlice: %v", err)
	}
	if string(jsonBytes) != "[]" {
		t.Fatalf("Esperado '[]', obtenido '%s'", string(jsonBytes))
	}

	// Test Scan con nil
	var s2 StringSlice
	if err := s2.Scan(nil); err != nil {
		t.Fatalf("Error escaneando nil: %v", err)
	}
	if s2 == nil || len(s2) != 0 {
		t.Fatalf("Esperado slice no-nil y longitud 0")
	}

	// Test Scan con "null" string
	var s3 StringSlice
	if err := s3.Scan("null"); err != nil {
		t.Fatalf("Error escaneando 'null': %v", err)
	}
	if s3 == nil || len(s3) != 0 {
		t.Fatalf("Esperado slice no-nil tras escanear 'null'")
	}

	// Test Scan y Value con elementos
	var s4 StringSlice
	if err := s4.Scan(`["584121234567","584129876543"]`); err != nil {
		t.Fatalf("Error escaneando array JSON: %v", err)
	}
	if len(s4) != 2 || s4[0] != "584121234567" {
		t.Fatalf("Elementos incorrectos tras scan: %v", s4)
	}

	val4, err := s4.Value()
	if err != nil {
		t.Fatalf("Error en Value: %v", err)
	}
	if val4 != `["584121234567","584129876543"]` {
		t.Fatalf("Valor serializado incorrecto: %v", val4)
	}
}
