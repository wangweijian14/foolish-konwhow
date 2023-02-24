package main

import (
	"naivete/ahocorasick"
	_ "naivete/internal/packed"

	_ "naivete/internal/logic"

	"github.com/gogf/gf/v2/os/gctx"

	"naivete/internal/cmd"
)

func main() {
	// dictionary := []string{
	// 	"mercury", "venus", "earth", "mars",
	// 	"jupiter", "saturn", "uranus", "pluto",
	// }
	// searches := _ahocorasick(dictionary, "we are earth ' man ! and not pluto's man")
	// matched := ""
	// for i := 0; i < len(searches); i++ {
	// 	matched += fmt.Sprintf("[%s]", dictionary[searches[i]])
	// }
	// fmt.Println("matched: ", matched)
	cmd.Main.Run(gctx.New())
}

func _ahocorasick(dictionary []string, match string) []int {

	m := ahocorasick.NewStringMatcher(dictionary)

	found := m.Match([]byte(match))

	return found
}
