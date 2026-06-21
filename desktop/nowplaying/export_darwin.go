//go:build desktop

package nowplaying

/*
#include <stdlib.h>
*/
import "C"

//export goNowPlayingCommand
func goNowPlayingCommand(cmd C.int, value C.double) {
	if globalHandler != nil {
		globalHandler(Command(cmd), float64(value))
	}
}
