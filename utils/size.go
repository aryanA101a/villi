package utils

import (
	"math"
	"strconv"
)

//by chatGPT
func ConvertToHumanReadable(size uint64) string {
	if size==0{return "0"}
	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	i := (math.Floor(math.Log(float64(size)) / math.Log(1024)))
	if i > 5 {
		i = 5
	}
	return strconv.FormatUint(uint64(size/uint64(math.Pow(1024, i))), 10) + sizes[int(i)]
}

