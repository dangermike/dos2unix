package main

import (
	"strings"
	"testing"

	_ "embed"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/1080-0.txt
var testText string

func TestEightCRLF(t *testing.T) {
	for _, test := range []string{
		"",
		"\r\n",
		"\r\r\n",
		"12345678",
		"\n2345678",
		"\r\n345678",
		"12\r\n5678",
		"123\r\n678",
		"1234\r\n78",
		"12345\r\n8",
		"123456\r\n",
		"1234567\r",

		"012345678",
		"\n12345678",
		"\r\n2345678",
		"0\r\n345678",
		"012\r\n5678",
		"0123\r\n678",
		"01234\r\n78",
		"012345\r\n8",
		"0123456\r\n",
		"01234567\r",

		"0000\n000000\r",
	} {
		t.Run(test, func(t *testing.T) {
			r := strings.NewReader(test)
			exp := strings.ReplaceAll(test, "\r\n", "\n")
			w := &strings.Builder{}
			n, err := dos2unix64(w, r)
			require.NoError(t, err)
			require.Equal(t, exp, w.String())
			require.Equal(t, len(exp), n)
		})
	}
}

type FakeWriter struct{}

func (w FakeWriter) Write(in []byte) (int, error) {
	return len(in), nil
}

func TestSourceFile(t *testing.T) {
	require.Equal(t, 39816, len(testText))
	var w FakeWriter
	r := strings.NewReader(testText)
	n, err := dos2unix64(w, r)
	require.NoError(t, err)
	require.Equal(t, 39091, n) // calculated using the "real" dos2unix tool
}

func BenchmarkDos2Unix(b *testing.B) {
	var w FakeWriter
	for n := 0; n < b.N; n++ {
		r := strings.NewReader(testText)
		_, _ = dos2unix64(w, r)
	}
}

func FuzzDos2Unix(f *testing.F) {
	f.Add("jackdaws love my big sphinx of quartz")
	f.Add("12\r\n5678")
	f.Add(testText)
	b := strings.Builder{}
	r := strings.NewReader("")
	f.Fuzz(func(t *testing.T, in string) {
		b.Reset()
		r.Reset(in)
		n, err := dos2unix64(&b, r)
		require.NoError(t, err)
		exp := strings.ReplaceAll(in, "\r\n", "\n")
		require.Equal(t, exp, b.String(), in)
		require.Equal(t, len(exp), n, in)
	})
}
