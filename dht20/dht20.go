package dht20

import (
	"fmt"
	"time"

	"github.com/d2r2/go-i2c"
)

const (
	defaultAddr = 0x38
	defaultBus  = 1

	initStatusReg = 0x71

	defaultInitDelay         = 100 * time.Millisecond
	defaultStartMeasureDelay = 10 * time.Millisecond
	defaultReadDataDelay     = 80 * time.Millisecond
)

type Conn interface {
	ReadRegU8(byte) (byte, error)
	WriteBytes([]byte) (int, error)
	ReadBytes([]byte) (int, error)
	Close() error
}

type DHT20 struct {
	i2c Conn

	initDelay         time.Duration
	startMeasureDelay time.Duration
	readDataDelay     time.Duration
}

func New() (*DHT20, error) {
	conn, err := i2c.NewI2C(defaultAddr, defaultBus)
	if err != nil {
		return nil, err
	}
	return NewWithConn(conn)
}

type Option func(*DHT20)

func WithDelays(initDelay, startMeasureDelay, readDataDelay time.Duration) Option {
	return func(d *DHT20) {
		d.initDelay = initDelay
		d.startMeasureDelay = startMeasureDelay
		d.readDataDelay = readDataDelay
	}
}

func NewWithConn(conn Conn, opts ...Option) (*DHT20, error) {
	sensor := &DHT20{
		i2c:               conn,
		initDelay:         defaultInitDelay,
		startMeasureDelay: defaultStartMeasureDelay,
		readDataDelay:     defaultReadDataDelay,
	}
	for _, opt := range opts {
		opt(sensor)
	}

	if err := sensor.init(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return sensor, nil
}

func (d *DHT20) Clean() {
	if d == nil || d.i2c == nil {
		return
	}
	_ = d.i2c.Close()
}

func (d *DHT20) init() error {
	if d == nil || d.i2c == nil {
		return fmt.Errorf("dht20: nil i2c connection")
	}

	time.Sleep(d.initDelay)
	_, err := d.i2c.ReadRegU8(initStatusReg)
	return err
}

func (d *DHT20) Get() (float64, float64, error) {
	if d == nil || d.i2c == nil {
		return 0, 0, fmt.Errorf("dht20: nil i2c connection")
	}

	time.Sleep(d.startMeasureDelay)
	if _, err := d.i2c.WriteBytes([]byte{0x00, 0xAC, 0x33, 0x00}); err != nil {
		return 0, 0, err
	}

	time.Sleep(d.readDataDelay)
	data := make([]byte, 7)
	if _, err := d.i2c.ReadBytes(data); err != nil {
		return 0, 0, err
	}

	humRaw := uint32(data[1])<<12 | uint32(data[2])<<4 | uint32(data[3]&0xF0)>>4
	tmpRaw := uint32(data[3]&0x0F)<<16 | uint32(data[4])<<8 | uint32(data[5])

	humidity := float64(humRaw) / 1048576.0 * 100.0
	temperature := float64(tmpRaw)/1048576.0*200.0 - 50.0

	return humidity, temperature, nil
}
