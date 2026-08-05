const { TRANSICIONES_CITA } = require('../src/config/constants');

/**
 * Tests unitarios para la máquina de estados de la Cita.
 * Verifica que las transiciones válidas e inválidas se manejan correctamente.
 */

describe('Máquina de Estados de Cita - Transiciones', () => {
  
  // ─── Transiciones válidas ────────────────────────────────
  describe('Transiciones válidas', () => {
    const casosValidos = [
      ['AGENDADA', 'ACEPTADA'],
      ['AGENDADA', 'CANCELADA'],
      ['ACEPTADA', 'EN_CURSO'],
      ['ACEPTADA', 'POSPUESTA'],
      ['ACEPTADA', 'CANCELADA'],
      ['EN_CURSO', 'PAUSADA'],
      ['EN_CURSO', 'FINALIZADA'],
      ['EN_CURSO', 'POSPUESTA'],
      ['EN_CURSO', 'CANCELADA'],
      ['PAUSADA', 'EN_CURSO'],
      ['PAUSADA', 'POSPUESTA'],
      ['PAUSADA', 'CANCELADA'],
      ['FINALIZADA', 'ENTREGADA'],
      ['POSPUESTA', 'ACEPTADA'],
    ];

    test.each(casosValidos)(
      '%s → %s debe ser válida',
      (estadoActual, nuevoEstado) => {
        const transiciones = TRANSICIONES_CITA[estadoActual];
        expect(transiciones).toBeDefined();
        expect(transiciones).toContain(nuevoEstado);
      }
    );
  });

  // ─── Transiciones inválidas ──────────────────────────────
  describe('Transiciones inválidas', () => {
    const casosInvalidos = [
      ['AGENDADA', 'EN_CURSO'],       // Debe pasar por ACEPTADA primero
      ['AGENDADA', 'FINALIZADA'],     // No se puede finalizar sin iniciar
      ['AGENDADA', 'ENTREGADA'],
      ['AGENDADA', 'POSPUESTA'],      // Solo técnico puede posponer tras aceptar
      ['ACEPTADA', 'FINALIZADA'],     // Debe pasar por EN_CURSO
      ['ACEPTADA', 'ENTREGADA'],
      ['EN_CURSO', 'ACEPTADA'],       // No puede retroceder
      ['EN_CURSO', 'AGENDADA'],
      ['EN_CURSO', 'ENTREGADA'],      // Debe pasar por FINALIZADA
      ['PAUSADA', 'FINALIZADA'],      // Debe reanudar primero
      ['PAUSADA', 'ACEPTADA'],
      ['FINALIZADA', 'EN_CURSO'],     // No puede retroceder
      ['FINALIZADA', 'CANCELADA'],    // Ya finalizada, no se cancela
      ['ENTREGADA', 'FINALIZADA'],    // Estado terminal
      ['ENTREGADA', 'CANCELADA'],
      ['ENTREGADA', 'EN_CURSO'],
      ['CANCELADA', 'AGENDADA'],      // Estado terminal
      ['CANCELADA', 'ACEPTADA'],
      ['CANCELADA', 'EN_CURSO'],
    ];

    test.each(casosInvalidos)(
      '%s → %s debe ser inválida',
      (estadoActual, nuevoEstado) => {
        const transiciones = TRANSICIONES_CITA[estadoActual];
        expect(transiciones).toBeDefined();
        expect(transiciones).not.toContain(nuevoEstado);
      }
    );
  });

  // ─── Estados terminales ──────────────────────────────────
  describe('Estados terminales', () => {
    test('ENTREGADA no tiene transiciones válidas', () => {
      expect(TRANSICIONES_CITA.ENTREGADA).toEqual([]);
    });

    test('CANCELADA no tiene transiciones válidas', () => {
      expect(TRANSICIONES_CITA.CANCELADA).toEqual([]);
    });
  });

  // ─── Reglas de negocio ─────────────────────────────────
  describe('Reglas de negocio', () => {
    test('POSPUESTA solo puede ir a ACEPTADA (reasignación)', () => {
      expect(TRANSICIONES_CITA.POSPUESTA).toEqual(['ACEPTADA']);
    });

    test('Solo desde estados activos se puede posponer', () => {
      // ACEPTADA, EN_CURSO, PAUSADA pueden ir a POSPUESTA
      expect(TRANSICIONES_CITA.ACEPTADA).toContain('POSPUESTA');
      expect(TRANSICIONES_CITA.EN_CURSO).toContain('POSPUESTA');
      expect(TRANSICIONES_CITA.PAUSADA).toContain('POSPUESTA');
      
      // AGENDADA, FINALIZADA, ENTREGADA NO pueden ir a POSPUESTA
      expect(TRANSICIONES_CITA.AGENDADA).not.toContain('POSPUESTA');
      expect(TRANSICIONES_CITA.FINALIZADA).not.toContain('POSPUESTA');
      expect(TRANSICIONES_CITA.ENTREGADA).not.toContain('POSPUESTA');
    });

    test('Solo desde estados activos se puede cancelar', () => {
      // AGENDADA, ACEPTADA, EN_CURSO, PAUSADA pueden ir a CANCELADA
      expect(TRANSICIONES_CITA.AGENDADA).toContain('CANCELADA');
      expect(TRANSICIONES_CITA.ACEPTADA).toContain('CANCELADA');
      expect(TRANSICIONES_CITA.EN_CURSO).toContain('CANCELADA');
      expect(TRANSICIONES_CITA.PAUSADA).toContain('CANCELADA');
      
      // FINALIZADA, ENTREGADA, POSPUESTA NO pueden ir a CANCELADA
      expect(TRANSICIONES_CITA.FINALIZADA).not.toContain('CANCELADA');
      expect(TRANSICIONES_CITA.ENTREGADA).not.toContain('CANCELADA');
      expect(TRANSICIONES_CITA.POSPUESTA).not.toContain('CANCELADA');
    });

    test('Todos los estados están definidos', () => {
      const estadosEsperados = [
        'AGENDADA', 'ACEPTADA', 'EN_CURSO', 'PAUSADA',
        'FINALIZADA', 'ENTREGADA', 'POSPUESTA', 'CANCELADA',
      ];
      for (const estado of estadosEsperados) {
        expect(TRANSICIONES_CITA[estado]).toBeDefined();
      }
    });
  });
});
