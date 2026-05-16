package words

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWordsService(t *testing.T) {
	testCase := []struct {
		request string
		actual  []string
	}{
		{
			request: "mines, captcha",
			actual:  []string{"mine", "captcha"},
		},
		{
			request: "",
			actual:  nil,
		},
		{
			request: "1, 2, 3",
			actual:  []string{"1", "2", "3"},
		},
		{
			request: "roman, Times",
			actual:  []string{"roman", "time"},
		},
	}
	for _, tc := range testCase {
		t.Run(tc.request, func(t *testing.T) {
			req := NormalizeTheWords(tc.request)
			fmt.Println(tc.actual, req)
			require.ElementsMatch(t, tc.actual, req)
		})
	}
}
