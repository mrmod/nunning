package buildtool

/*
	Input

Implements Stringable
Can derive a String() => string from Input
*/
type Input string

func (i Input) String() string {
	return string(i)
}
