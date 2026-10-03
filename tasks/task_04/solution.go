package main

type Stats struct {
	Count         int
	Sum, Min, Max int64
}

func Calc(nums []int64) Stats {
	if len(nums) < 2 {
		return Stats{}
	}

	var sum int64 = 0

	min := nums[1] - nums[0]
	max := nums[1] - nums[0]
	for i := 1; i < len(nums); i++ {
		tmp_res := nums[i] - nums[i-1]

		sum += tmp_res

		if min > tmp_res {
			min = tmp_res
		}

		if max < tmp_res {
			max = tmp_res
		}
	}

	return Stats{Count: len(nums) - 1, Sum: int64(sum), Min: min, Max: max}

}
