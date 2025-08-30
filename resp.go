package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

type Value struct {
	typ   string
	str   string
	num   int
	bulk  string
	array []Value
}

type Resp struct {
	reader *bufio.Reader
}

func NewResp(conn net.Conn) *Resp {
	return &Resp{
		reader: bufio.NewReader(conn),
	}
}

func (r *Resp) Read() (Value, error) {
	b, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch b {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		return Value{}, fmt.Errorf("unknown type: %c", b)
	}
}

func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' && line[len(line)-1] == '\n' {
			return line[:len(line)-2], n, nil
		}
	}
}

func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, n, err
	}

	i, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}

	return int(i), n, nil
}

func (r *Resp) readArray() (Value, error) {
	val := Value{typ: "array"}

	length, _, err := r.readInteger()
	if err != nil {
		return val, err
	}

	val.array = make([]Value, length)
	for i := 0; i < length; i++ {
		v, err := r.Read()
		if err != nil {
			return val, err
		}
		val.array[i] = v
	}

	return val, nil
}

func (r *Resp) readBulk() (Value, error) {
	val := Value{typ: "bulk"}

	length, _, err := r.readInteger()
	if err != nil {
		return val, err
	}
	bulk := make([]byte, length)
	r.reader.Read(bulk)
	val.bulk = string(bulk)
	r.readLine()

	return val, nil
}
