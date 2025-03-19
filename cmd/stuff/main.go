package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
)

type path []byte

func add(i int, j int) (int, error) { return i + j, nil }

func sub(i int, j int) (int, error) { return i - j, nil }

func mul(i int, j int) (int, error) { return i * j, nil }

func div(i int, j int) (int, error) {
	if j == 0 {
		return 0, errors.New("division by zero")
	}
	return i / j, nil
}

var opMap = map[string]func(int, int) (int, error){
	"+": add,
	"-": sub,
	"*": mul,
	"/": div,
}

func fileLen(file string) (int64, error) {
	f, err := os.Open(file)
	if err != nil {
		return 0, fmt.Errorf("err: %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return 0, fmt.Errorf("problemn reading stat of file: %s. err: %w", file, err)
	}
	return stat.Size(), nil
}

func prefixer(prefix string) func(string) string {
	return func(input string) string {
		return fmt.Sprintf("%s %s", prefix, input)
	}
}

type Person struct {
	FirstName, LastName string
	Age                 int
}

func (p Person) String() string {
	return fmt.Sprintf("firstName: %s, lastName: %s, age: %d", p.FirstName, p.LastName, p.Age)
}

func MakePerson(firstName, lastName string, age int) Person {
	return Person{
		FirstName: firstName,
		LastName:  lastName,
		Age:       age,
	}
}

func MakePersonPointer(firstName, lastName string, age int) *Person {
	return &Person{
		FirstName: firstName,
		LastName:  lastName,
		Age:       age,
	}
}

func UpdateSlice(s []string, str string) {
	s[len(s)-1] = str
	fmt.Println(s)
}

func GrowSlice(s []string, str string) {
	s = append(s, str)
	fmt.Println(s)
}

type teamName string
type player string

type Team struct {
	name    teamName
	players []player
}

type League struct {
	Teams []Team
	Wins  map[teamName]int
}

func NewLeague() *League {
	return &League{Wins: make(map[teamName]int)}
}

func (l *League) MatchResult(team1, team2 teamName, score1, score2 int) {
	switch r := score1 - score2; {
	case r > 0:
		l.Wins[team1] += 1
	case r < 0:
		l.Wins[team2] += 1
	}
}

func (l *League) Ranking() []teamName {
	names := make([]teamName, 0, len(l.Teams))
	for _, t := range l.Teams {
		names = append(names, t.name)
	}
	sort.Slice(names, func(i, j int) bool {
		return l.Wins[names[i]] > l.Wins[names[j]]
	})
	return names
}

type Ranker interface {
	Ranking() []string
}

func RankPrinter(r Ranker, w io.Writer) {
	for _, v := range r.Ranking() {
		io.WriteString(w, v+"\n")
	}
}

func Filter[T any](s []T, f func(T) bool) []T {
	var r []T
	for _, v := range s {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

func Convert[T1, T2 Integer](in T1) T2 {
	return T2(in)
}

type Integer interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

type IntegerFloat interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

func IntegerDouble[T IntegerFloat](in T) T {
	return in * 2
}

type Printable[T any] interface {
	~int | ~float64
	fmt.Stringer
}

type MyInt int
type MyFloat float64

func (m MyInt) String() string {
	return fmt.Sprintf("MyInt is %d\n", m)
}

func (m MyFloat) String() string {
	return fmt.Sprintf("MyFloat is %f\n", m)
}

func Print[T Printable[T]](p T) {
	fmt.Println(p)
}

type LinkedList[T comparable] struct {
	root   *Node[T]
	length int
}

type Node[T comparable] struct {
	val  T
	next *Node[T]
}

func (n *Node[T]) String() string {
	return fmt.Sprintf("Node value: %v. Next node: %+v", n.val, n.next)
}

func (n *LinkedList[T]) Add(v T) {
	defer func() { n.length++ }()

	if n.root == nil {
		n.root = &Node[T]{val: v}
		return
	}

	var current = n.root
	for current != nil {
		if current.next == nil {
			break
		}
		current = current.next
	}

	// create a new Node
	nn := &Node[T]{val: v}
	current.next = nn

	// fmt.Println("root " + n.root.String())
}

func (n *LinkedList[T]) Insert(v T, i int) error {
	if i > n.length {
		return fmt.Errorf("out of bound index: %d. Length is: %d", i, n.length)
	}

	//
	nn := &Node[T]{val: v}
	if n.length == 0 {
		n.root = nn
	}

	c := n.root
	for j := 0; j < i-1; j++ {
		c = c.next
	}

	if i == 0 {
		nn.next = n.root
		n.root = nn
		n.length++
		return nil
	}

	nn.next = c.next
	c.next = nn
	n.length++
	return nil
}

func (n *LinkedList[T]) Index(v T) int {
	if n.length == 0 {
		return -1
	}

	c := n.root
	for i := 0; i < n.length; i++ {
		if c.val == v {
			return i
		}
		c = c.next
	}
	return -1
}

func NewLinkedList[T comparable](in T) *LinkedList[T] {
	return &LinkedList[T]{}
}

func main() {
	//
	// CHAPTER 8
	//
	ll := NewLinkedList(5)
	ll.Add(10)
	ll.Add(20)
	ll.Add(30)
	ll.Add(40)
	// fmt.Printf("Length is: %d\n", ll.length)
	if err := ll.Insert(50, 0); err != nil {
		fmt.Println(err)
	}
	if err := ll.Insert(100, 2); err != nil {
		fmt.Println(err)
	}
	if err := ll.Insert(1000, 6); err != nil {
		fmt.Println(err)
	}
	fmt.Println("What is the index of 10?: ", ll.Index(50))
	// fmt.Printf("%+v\n", *ll.root)

	for l := ll; l.root != nil; l.root = l.root.next {
		fmt.Println(l.root.val)
	}

	// Print(MyInt(5))
	// Print(MyFloat(10.1111))

	// fmt.Println(IntegerDouble(10.15))
	// fmt.Println(IntegerDouble())

	// END

	// var a int = 10
	// b := Convert[int, int64](a)
	// rb := reflect.TypeOf(b).Kind()
	// fmt.Printf("%+v\n", rb)

	// words := []string{"One", "Potato", "Two", "Potato"}

	// filtered := Filter(words, func(s string) bool {
	// 	return s != "Potato"
	// })
	// fmt.Printf("len: %d, filtered: %v\n", len(filtered), filtered)
	// league := League{
	// 	Teams: []Team{
	// 		{name: "test", players: []player{"joro"}},
	// 		{name: "test2", players: []player{"tony"}},
	// 	},
	// 	Wins: map[teamName]int{
	// 		"test":  3,
	// 		"test2": 2,
	// 	},
	// }
	// fmt.Println(league)
	// fmt.Println(league.Ranking())
}

// func main() {
// fmt.Println(fileLen("non_existing"))
// helloPrefix := prefixer("Hello")
// fmt.Println(helloPrefix("Bob"))   // should print Hello Bob
// fmt.Println(helloPrefix("Maria")) // should print Hello Maria
//
//
// CHAPTER 6
//
// mp := MakePerson("George", "Yanev", 41)
// mpp := MakePersonPointer("Tony", "Yaneva", 42)
// fmt.Println("MakePerson is: " + mp.String())
// fmt.Println("MakePersonPointer is: " + mpp.String())
// MakePerson("test", "ttt", 3)
//

// s := []string{"update"}
// fmt.Println("before slice being updated: " + strings.Join(s, ""))
// UpdateSlice(s, "test2")
// fmt.Println("after slice being updated: " + strings.Join(s, ""))

// sGrow := make([]string, 1, 2)
// sGrow[0] = "grow"

// fmt.Println("before adding new element to the slice: " + strings.Join(sGrow, ""))
// GrowSlice(sGrow, "grow2")
// fmt.Println("after adding new element to the slice: " + strings.Join(sGrow, ""))
//

// }

// func main() {
// 	expressions := [][]string{
// 		{"2", "+", "3"},
// 		{"2", "-", "3"},
// 		{"2", "*", "3"},
// 		{"2", "/", "3"},
// 		{"2", "/", "0"},
// 		{"2", "%", "3"},
// 		{"two", "+", "three"},
// 		{"5"},
// 	}
// 	for _, expression := range expressions {
// 		if len(expression) != 3 {
// 			fmt.Println("invalid expression:", expression)
// 			continue
// 		}
// 		p1, err := strconv.Atoi(expression[0])
// 		if err != nil {
// 			fmt.Println(err)
// 			continue
// 		}
// 		op := expression[1]
// 		opFunc, ok := opMap[op]
// 		if !ok {
// 			fmt.Println("unsupported operator:", op)
// 			continue
// 		}
// 		p2, err := strconv.Atoi(expression[2])
// 		if err != nil {
// 			fmt.Println(err)
// 			continue
// 		}
// 		result, err := opFunc(p1, p2)
// 		if err != nil {
// 			fmt.Printf("operation: %s, par1: %v, par2: %v\n", op, p1, p2)
// 		}
// 		fmt.Println(result)
// 	}
// }

// func main() {
// s := []int{3, 4, 2, 15, 0, 100}
// ShellSort(s)
// fmt.Println(s)

// var b byte
// var smallI int32 = (1 << 31) - 1
// var bigI uint64 = (1 << 64) - 1
// fmt.Println(smallI + 1)
// fmt.Printf("float %d\n", bigI)

// x := make([]string, 0, 5)
// x = append(x, "a", "b", "c", "d")
// y := x[:2]
// z := x[2:]
// fmt.Println(cap(x), cap(y), cap(z)) // 5, 5, 3
// y = append(y, "i", "j", "k")        // "a", "b", "i", "j", "k"
// x = append(x, "x")                  // "a", "b", "i", "j", "k", "x"
// z = append(z, "y")                  // "i", "j", "k", "x", "y"
// fmt.Println("x:", x)
// fmt.Println("y:", y)
// fmt.Println("z:", z)

// x := []int{1, 2, 3, 4}
// d := [4]int{5, 6, 7, 8}
// y := make([]int, 2)
// copy(y, d[:])
// fmt.Println(y)
// copy(d[:], x)
// fmt.Println(d)

//
// CHAPTER 3
//

// greetings := []string{"Hello", "Hola", "नमस्कार", "こんにちは", "Привіт"}
// firstSlice := greetings[:2]
// secondSlice := greetings[1:4]
// thirdSlice := greetings[3:]
// fmt.Println("first slice", firstSlice)
// fmt.Println("second slice", secondSlice)
// fmt.Println("third slice", thirdSlice)

// message := "Hi 😱 and 😁"
// b := []byte(message)
// r := bytes.Runes(b)
// fmt.Printf("message as rune: %q\n", r[3])

// type Employee struct {
// 	firstName, lastName string
// 	id                  int
// }

// e1 := Employee{
// 	"First1",
// 	"Last1",
// 	1,
// }

// e2 := Employee{
// 	firstName: "First2",
// 	lastName:  "Last2",
// 	id:        2,
// }

// var e3 Employee
// e3.firstName = "First3"
// e3.lastName = "Last3"
// e3.id = 3

// fmt.Printf("first: %v, second: %v, third: %v\n", e1, e2, e3)

//
// CHAPTER 4
//

// var random []int
// for i := 0; i < 100; i++ {
// 	random = append(random, rand.Intn(101))
// }
// for _, i := range random {
// 	switch {
// 	case i%2 == 0 && i%3 == 0:
// 		fmt.Println("Divide by six: ", i)
// 	case i%2 == 0:
// 		fmt.Println("Divide by two: ", i)
// 	case i%3 == 0:
// 		fmt.Println("Divide by three: ", i)
// 	}
// }
// fmt.Printf("random length: %d\nvalue: %v\n", len(random), random)

// var total int
// for i := 0; i < 10; i++ {
// 	total := total + i
// 	fmt.Println(total)
// }
// fmt.Println(total)

//
// CHAPTER 5
//

// fmt.Println(addTo(3))
// fmt.Println(addTo(3, 2))

// expressions := [][]string{
// 	{"2", "+", "3"},
// 	{"2", "-", "3"},
// 	{"2", "*", "3"},
// 	{"2", "/", "3"},
// 	{"2", "%", "3"},
// 	{"two", "+", "three"},
// 	{"5"},
// }

// for _, e := range expressions {
// 	fmt.Println(e)
// }

// }

func addTo(base int, vals ...int) []int {
	out := make([]int, 0, len(vals))
	for _, v := range vals {
		out = append(out, base+v)
	}
	return out
}

func InsertionSort(s []int) {
	for i := 1; i < len(s); i++ {
		c := s[i]
		j := i - 1
		for j >= 0 && s[j] > c {
			s[j+1] = s[j]
			j--
		}
		s[j+1] = c
	}
}

func ShellSort(s []int) {
	len := len(s)
	if len <= 1 {
		return
	}

	for gap := len / 2; gap > 0; gap /= 2 {
		for i := gap; i < len; i++ {
			c := s[i]
			j := i
			for j >= gap && s[j-gap] > c {
				s[j] = s[j-gap]
				j -= gap
			}
			s[j] = c
		}

	}
}
