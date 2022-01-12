package ztype

import (
	"fmt"
	"github.com/protolambda/ztyp/tree"
)

func sum(nums ...int) {
	fmt.Print(nums, " ")
	total := 0
	for _, num := range nums {
		total += num
	}
	fmt.Println(total)
}

func ConcatGeneralizedIndices(indices ...tree.Gindex64) tree.Gindex64 {
	o := tree.Gindex64(1)
	for _, i := range indices {
		o = o*GetPowerOfTwoFloor(i) + (i - GetPowerOfTwoFloor(i))
	}
	return o
}

/*
def get_power_of_two_ceil(x: int) -> int:
    """
    Get the power of 2 for given input, or the closest higher power of 2 if the input is not a power of 2.
    Commonly used for "how many nodes do I need for a bottom tree layer fitting x elements?"
    Example: 0->1, 1->1, 2->2, 3->4, 4->4, 5->8, 6->8, 7->8, 8->8, 9->16.
    """
    if x <= 1:
        return 1
    elif x == 2:
        return 2
    else:
        return 2 * get_power_of_two_ceil((x + 1) // 2)
*/

func GetPowerOfTwoCeil(x tree.Gindex64) tree.Gindex64 {
	if x <= 1 {
		return 1
	} else if x == 2 {
		return 2
	} else {
		return 2 * GetPowerOfTwoCeil((x+1)/2)
	}
}

/*
def get_power_of_two_floor(x: int) -> int:
    """
    Get the power of 2 for given input, or the closest lower power of 2 if the input is not a power of 2.
    The zero case is a placeholder and not used for math with generalized indices.
    Commonly used for "what power of two makes up the root bit of the generalized index?"
    Example: 0->1, 1->1, 2->2, 3->2, 4->4, 5->4, 6->4, 7->4, 8->8, 9->8
    """
    if x <= 1:
        return 1
    if x == 2:
        return x
    else:
        return 2 * get_power_of_two_floor(x // 2)
*/

func GetPowerOfTwoFloor(x tree.Gindex64) tree.Gindex64 {
	if x <= 1 {
		return 1
	} else if x == 2 {
		return 2
	} else {
		return 2 * GetPowerOfTwoFloor(x/2)
	}
}
