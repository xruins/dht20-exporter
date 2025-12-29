package dht20

import (
	"math"
	"testing"
	"time"
)

type mockConn struct {
	initRegRead bool
	wroteStart  bool
	readData    []byte
}

func (m *mockConn) ReadRegU8(reg byte) (byte, error) {
	if reg != initStatusReg {
		return 0, &unexpectedCallError{msg: "ReadRegU8 unexpected reg"}
	}
	m.initRegRead = true
	return 0x1c, nil
}

func (m *mockConn) WriteBytes(b []byte) (int, error) {
	want := []byte{0x00, 0xAC, 0x33, 0x00}
	if len(b) != len(want) {
		return 0, &unexpectedCallError{msg: "WriteBytes unexpected length"}
	}
	for i := range b {
		if b[i] != want[i] {
			return 0, &unexpectedCallError{msg: "WriteBytes unexpected payload"}
		}
	}
	m.wroteStart = true
	return len(b), nil
}

func (m *mockConn) ReadBytes(buf []byte) (int, error) {
	if len(buf) != 7 {
		return 0, &unexpectedCallError{msg: "ReadBytes unexpected length"}
	}
	if len(m.readData) != 7 {
		return 0, &unexpectedCallError{msg: "ReadBytes missing fixture"}
	}
	copy(buf, m.readData)
	return len(buf), nil
}

func (m *mockConn) Close() error { return nil }

type unexpectedCallError struct{ msg string }

func (e *unexpectedCallError) Error() string { return e.msg }

func encodeMeasurement(humidityPct, temperatureC float64) []byte {
	humRaw := uint32(math.Round(humidityPct / 100.0 * 1048576.0))
	tmpRaw := uint32(math.Round((temperatureC + 50.0) / 200.0 * 1048576.0))

	data := make([]byte, 7)
	data[1] = byte(humRaw >> 12)
	data[2] = byte(humRaw >> 4)
	data[3] = byte((humRaw&0x0F)<<4) | byte((tmpRaw>>16)&0x0F)
	data[4] = byte(tmpRaw >> 8)
	data[5] = byte(tmpRaw)
	return data
}

func TestGetParsesHumidityAndTemperature(t *testing.T) {
	conn := &mockConn{
		readData: encodeMeasurement(55.5, 22.25),
	}

	sensor, err := NewWithConn(conn, WithDelays(0, 0, 0))
	if err != nil {
		t.Fatalf("NewWithConn: %v", err)
	}
	t.Cleanup(sensor.Clean)

	hum, tmp, err := sensor.Get()
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if !conn.initRegRead {
		t.Fatalf("expected init reg read")
	}
	if !conn.wroteStart {
		t.Fatalf("expected start-measure command written")
	}

	if math.Abs(hum-55.5) > 0.01 {
		t.Fatalf("humidity = %v, want ~%v", hum, 55.5)
	}
	if math.Abs(tmp-22.25) > 0.01 {
		t.Fatalf("temperature = %v, want ~%v", tmp, 22.25)
	}
}

func TestWithDelaysSetsAllDelays(t *testing.T) {
	conn := &mockConn{readData: make([]byte, 7)}
	sensor, err := NewWithConn(conn, WithDelays(1*time.Millisecond, 2*time.Millisecond, 3*time.Millisecond))
	if err != nil {
		t.Fatalf("NewWithConn: %v", err)
	}
	if sensor.initDelay != 1*time.Millisecond || sensor.startMeasureDelay != 2*time.Millisecond || sensor.readDataDelay != 3*time.Millisecond {
		t.Fatalf("delays not set: init=%v start=%v read=%v", sensor.initDelay, sensor.startMeasureDelay, sensor.readDataDelay)
	}
}

