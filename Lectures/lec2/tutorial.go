package tutorial

import (
	"fmt"
	"math"
)

/* slice */
func slice() {
	// An array's length is part of its type, so arrays cannot be resized.
	// A slice is a dynamically-sized flexible view into the array.
	arr := [4]int{1, 2, 3, 4}
	fmt.Println(arr)

	var slice []int = arr[:]
	fmt.Println(slice)

	slice = arr[1:3]
	fmt.Println(slice)

	slice = slice[0:1]
	fmt.Println(slice)

	// A slice doesn't store any data, but describes a section of the array.
	// Slices are like references to arrays. Changing elements of slices
	// modifies the elements in the underlying array. Other slices sharing the
	// same array would see the changes.

	// A slice literal is like an array literal without the length.
	// [3]bool{true, true, false} is an array literal.
	// []bool{true, true, false}  creates the same array as above, then builds
	// a slice that references it.

	// A slice has a length and a capacity.
	// The length is the number of elements it contains.
	// The capacity is the number of elements in the underlying array, counting
	// from the first element in the slice.
	// You can extend a slice's length by reslicing it.
	fmt.Println(len(slice), cap(slice))

	// The zero value for a slice is nil.
	var sliceZeroVal []int
	fmt.Println(sliceZeroVal, len(sliceZeroVal), cap(sliceZeroVal))

	// Slice can be created using `make` function. This creates
	// dynamically-sized arrays. The array is zeroed.
	makeSlice := make([]int, 5)
	fmt.Println(makeSlice)

	// Slice can contain any type, including other slices.
	board := [][]string{
		[]string{"_", "_", "_"},
		[]string{"_", "_", "_"},
		[]string{"_", "_", "_"},
	}
	board[0][0] = "X"

	// Append to a slice
	// func append(s []T, v1 T, v2 T, ...)
	// The return value of `append` is a slice containing elements of the
	// original slice plus the provided values.
	// If the backing array of the slice is too small, more space is allocated.
	// The returned slice would point to the newly allocated array.
	var s []int
	fmt.Println(s, len(s), cap(s))

	s = append(s, 0)
	fmt.Println(s, len(s), cap(s))

	s = append(s, 1, 2)
	fmt.Println(s, len(s), cap(s))

	// The `range` form of the `for` loop iteratives over a slice or a map.
	// When ranging over a slice, two values are returned for each iteration.
	// The first is the index, the second is a COPY of the element at the index.
	// If you only want the idnex, you can omit the second variable.
	pow := make([]int, 10)
	for i := range pow {
		pow[i] = 1 << uint(i) // == 2**i
	}
	for _, value := range pow {
		fmt.Printf("%d\n", value)
	}
}

/* map */
func maps() {
	// The zero value of a map is nil.
	// Use `make` to get a map.
	m := make(map[string]int)
	m["One"] = 0
	fmt.Println(m)

	// Mutating maps
	// insert or update
	m["One"] = 1
	m["Two"] = 2

	// retrieve
	fmt.Println(m["Two"])

	// delete
	delete(m, "Two")

	// test existence
	two, ok := m["Two"]
	if ok {
		fmt.Printf("Delete fail! Two: %d\n", two)
	} else {
		fmt.Println("Delete succeeds!")
	}
}

/* function */
func add(x int, y int) int {
	return x + y
}

func compute(fn func(int, int) int, a int, b int) int {
	return fn(a, b)
}

func adder() func(int) int {
	sum := 0
	return func(x int) int {
		sum += x
		return sum
	}
}

func funcs() {
	// In Go, functions are values and can be used as argments, return values.'
	fmt.Println(compute(add, 40, 2))

	// Function closure
	// Go functions can be closures. A closure is a function value that
	// uses variables outside its body. The function is "bound" to the variable.
	// In the example, each closure is bound to its OWN `sum` variable.
	pos, neg := adder(), adder()
	for i := 0; i < 10; i++ {
		fmt.Println(
			pos(i),
			neg(-2*i),
		)
	}
}

/* method */
type Vertex struct {
	X float64
	Y float64
}

// Go doesn't have classes, but we can define methods on types. A method is
// a function with a special receiver argument.
// A method is just a normal function, but with a receiver. You can re-write
// it as a normal function, with no change in functionality.
func (v Vertex) abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func abs(v Vertex) float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// You can delcare methods on non-struct types. You can only define method
// with a receiver whose type is defined in the same package as the method.
// So you cannot define a method for `float64` directly.
type MyFloat float64

func (f MyFloat) abs() float64 {
	if f < 0 {
		return float64(-f)
	}
	return float64(f)
}

// You can declare methods with pointer receivers.
// Methods with pointer receivers can modify the value pointed by the receiver.
// As methods often need to modify the receiver, pointer receivers are more
// common than value receivers.
// Note: Go always passes by value. So scaleWrong() doesn't work even if called
// as v.scaleWrong(10). This is syntactic sugar for scaleWrong(v, 10). If a
// pointer is the argument, a copy of the pointer is passed.
func (v *Vertex) scale(f float64) {
	v.X *= f
	v.Y *= f
}

func (v Vertex) scaleWrong(f float64) {
	v.X *= f
	v.Y *= f
}

// 1. Functions with a pointer argument must take a pointer, but methods with
// pointer receivers take either a value or a pointer as the receiver. For
// method with pointer receiver, if a value is supplied, Go automatically
// turns the value into a pointer.
// 2. Functions taking a value argument must take a value, while methods with
// value receivers take either a value or a pointer as the value. Go
// automatically dereferences the pointer if needed.

// Reason for using pointer receiver: (1) modify the value; (2) avoid copying.
// In general, use either one, but not a mixure.
func methods() {
	v := Vertex{3, 4}
	fmt.Println(v.abs())
	fmt.Println(abs(v))

	f := MyFloat(-math.Sqrt2)
	fmt.Println(f.abs())

	v.scaleWrong(10)
	fmt.Println(v.abs())

	v.scale(10)
	fmt.Println(v.abs())
}

/* interface */

// An interface type is a set of method signatures.
// A value of interface type can hold any value that implements those methods.

// A type implements an interface by implementing its methods, but no explicit
// declaration is needed. This's by design.

// A interface value can be thought of as a (value, type) tuple, holding a
// value of the concrete type. Calling a method on an interface value
// executes the method of the same name on the underlying type.
type printable interface {
	print()
}

func (v *Vertex) print() {
	if v == nil {
		fmt.Println("<nil>")
		return
	}
	fmt.Printf("Vertex{X: %f, Y: %f}\n", v.X, v.Y)
}

// The empty interface can hold values of any type. This's ususally used by
// code that handles unknown type. For example, fmt.Print.

func interfaces() {
	var p printable

	// If the concrete value inside the interface value is nil, the method
	// would be called with a nil receiver. This's not an error in Go.
	// An interface value holding a nil concrete value is non-nil. It holds
	// nil concrete value but non-nil concrete type.
	var v *Vertex

	p = v
	p.print()

	p = &Vertex{3, 4}
	p.print()

	// A nil interface value holds neither value nor concrete type. Calling
	// a method on a nil interface is a run-time error, because there's no
	// type inside the interface tuple to indicate which concrete method to
	// call.
	var nilInterfaceValue printable
	nilInterfaceValue.print() // runtime error!

	// Type assertions
	// A type assertion privides access to an interface value's underlying
	// concrete value.
	// t := i.(T)
	// asserts that the interface value i holds concrete type T and assigns
	// the underlying T value to the variable t. If this is false, Go panics.
	// t, ok := i.(T)
	// If i holds T, then ok would be true, and t holds the value. If not, ok
	// is false and t is the zero value of T.

	// Type switches
	// Like a regular switch statement, but allows `case` to be types.
	// switch v := i.(type) {
	// case T:
	//     // here v has type T
	// case S:
	//     // here v has type S
	// default:
	//     // no match; here v has the same type as i
	// }
	// In i.(type), `type` is a reserved keyword.
}

/* Errors */
// Go expresses error state with `error` values. `error` is a built-in type.
// type error interface {
//     Error() string
// }
// The caller of functions returning error should handler errors by testing
// whether the error equals nil.
//
// i, err := strconv.Atoi("42")
// if err != nil {
//     fmt.Printf("couldn't convert: %v\n", err)
//     return
// }
// fmt.Println("Converted integer:", i)

/* concurrency */
func concurrency() {
	// A goroutine is a lightweight thread managed by the Go runtime.
	// go f(x, y, z)
	// starts a new goroutine. The evaluation of x, y, z happens in the current
	// goroutine and the execution of f happens in the new goroutine.
	// Goroutines run in the same address space, so synchronization is needed.

	// Channels
	// Send and receiver with <- operator. Must be created before use.
	// By default, sends and receives are blocking.
	// ch := make(chan int)

	// Channels can be buffered
	// bufferCh := make(chan int, 100)

	// A sender can close a channel. Receiver can test whether a channel is
	// closed by assigning a second parameter to the receive expression.
	// v, ok := <- ch

	// The `range` form of the for loop receives values from the channel
	// repeatedly until it's closed.
	// ONLY the sender should close the channel, never the receiver. Sending
	// on a closed channel would cause a panic! Channels, different from files,
	// is not required to be closed. Closing is needed to tell the receiver
	// that no one is sending.

	// A select statement lets a goroutine wait on multiple channels. The
	// default case runs if no other case is ready. This avoids blocking.
}

func main() {
	// slice()
	// maps()
	// funcs()
	// methods()
	// interfaces()
}
