package main

func rotateRunes(s string, shift int) string {
	rune_slice := []rune(s)

	n := len(rune_slice)

	if n <= 1 {
		return string(rune_slice)
	}

	shift = shift % n

	if shift < 0 {
		shift = shift + n
	}

	if shift == 0 {
		return string(rune_slice)
	}

	res := append(rune_slice[shift:], rune_slice[:shift]...)
	return string(res)
}

//Не эффективное решение, не прошло по времени
// func rotateRunes(s string, shift int) string {
// 	if strings.TrimSpace(s) == "" {
// 		return s
// 	}

// 	rune_slice := []rune(s)

// 	n := len(rune_slice)

// 	shift = shift % n

// 	if shift < 0 {
// 		for j := 0; j < max(shift, -shift); j++ {
// 			tmp := rune_slice[n-1]
// 			for i := n - 1; i > 0; i-- {
// 				rune_slice[i] = rune_slice[i-1]
// 			}
// 			rune_slice[0] = tmp
// 		}
// 		return string(rune_slice)
// 	}

// 	for j := 0; j < shift; j++ {
// 		tmp := rune_slice[0]
// 		for i := 0; i < n-1; i++ {
// 			rune_slice[i] = rune_slice[i+1]
// 		}
// 		rune_slice[n-1] = tmp
// 	}

// 	return string(rune_slice)
// }
