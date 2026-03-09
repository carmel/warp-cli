package utils

import (
	"errors"

	"github.com/carmel/warp-cli/wireguard/conn"
)

type Bind struct {
	conn.Bind
	reseved [3]byte
}

func NewResevedBind(bind conn.Bind, reserved [3]byte) *Bind {
	return &Bind{
		Bind:    bind,
		reseved: reserved,
	}
}

func (b *Bind) SetReseved(reserved [3]byte) {
	b.reseved = reserved
}

func (b *Bind) Send(buf []byte, ep conn.Endpoint) error {
	if len(buf) > 3 {
		buf[1] = b.reseved[0]
		buf[2] = b.reseved[1]
		buf[3] = b.reseved[2]
	}

	// Wrap buf in a slice of slices to match [][]byte
	return b.Bind.Send([][]byte{buf}, ep)
}

func (b *Bind) Open(port uint16) (fns []conn.ReceiveFunc, actualPort uint16, err error) {
	fns, actualPort, err = b.Bind.Open(port)
	if err != nil {
		return
	}

	var tempFns []conn.ReceiveFunc
	for _, fn := range fns {
		tempFns = append(tempFns, b.NewReceiveFunc(fn))
	}

	fns = tempFns
	return
}

func (b *Bind) NewReceiveFunc(fn conn.ReceiveFunc) conn.ReceiveFunc {
	return func(bufs [][]byte, sizes []int, eps []conn.Endpoint) (int, error) {
		n, err := fn(bufs, sizes, eps)
		if err != nil || n == 0 {
			return n, err
		}
		for i := 0; i < n; i++ {
			if sizes[i] < 4 {
				return n, errors.New("buffer too small")
			}
			if bufs[i][1] != b.reseved[0] || bufs[i][2] != b.reseved[1] || bufs[i][3] != b.reseved[2] {
				return n, errors.New("bad reseved")
			}
			bufs[i][1] = 0
			bufs[i][2] = 0
			bufs[i][3] = 0
		}
		return n, nil
	}
}
