package main

import "math/rand"
import "time"
import term "github.com/buger/goterm"
import "syscall"
import "os"
import "os/signal"



// TODO: Add retro text UI, such as:
//   ┌──[ Go Life ]─────────────────────────────────────┐
//   │..................................................│
//   │..................................................│
//   │..................................................│
//   │..................................................│
//   │..................................................│
//   │..................................................│
//   │..................................................│
//   │..................................................│
//   │..................................................│
//   │..................................................│
//   │..................................................│
//   │..................................................│
//   │..................................................│
//   │..................................................│
//   │..................................................│
//   │..................................................│
//   └──────────────────────────────────────────────────┘
//    [wrap] [pause] [quit]   Cells: xxxxx  Tick: xxxxxxx

// type TUI_button struct {
//   x         int
//   y         int
//   tabindex  int
//   text      string
//   color     string
// }

// type TUI struct {
//   width     int
//   height    int
//   tabindex  int
//   controls  *[]TUI_button
// }



type State struct {
    world   *World
    paused  bool
    fps     int
}

type World struct {
    width   int
    height  int
    cells   *[]bool
    tick    uint
}


// Aviators: Building better worlds together
func newWorld(width, height int) *World {
    cells := make([]bool, width*height)
    return &World{ width, height, &cells, 0 }
}

// Randomize world cells
func randomizeWorld(world *World) {
    rand.Seed( time.Now().UnixNano() )
    for i:=0; i<cap(*world.cells); i++ {
        (*world.cells)[i] = (rand.Intn(10) > 4)
    }
}

// Handle interruptions
func initInterruptHandler(state *State) {
    channel := make(chan os.Signal)
    signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
    go func() {
        <- channel
        state.paused = true
        time.Sleep(time.Second/8)
        term.Clear()
        term.MoveCursor(1,1)
        term.Flush()
        defer os.Exit(0)
    }()
}


func main() {
    // Determine available screen estate
    screenHeight := term.Height()
    screenWidth  := term.Width()
    worldHeight  := screenHeight-1  // no scrolling
    worldWidth   := screenWidth

    // Create our world
    var world *World = newWorld(worldWidth, worldHeight)
    randomizeWorld(world)  // and fill it with something interesting

    // Keep track of state
    var state *State = &State{world, false /*paused*/, 30 /*fps*/}

    // Gracefully handle interruptions instead of yelling
    initInterruptHandler(state)

    // and churn it
    mainLoop(state)
}


func mainLoop(state *State) {
    var nextTick, now int64
    for {
        // FPS limiting, bitches!
        nextTick = time.Now().UnixNano() + (int64(time.Second) / int64(state.fps))

        // Update and draw the world
        if (state.paused == false) {
            state.world = tickWorld(state.world)
            drawWorld(state.world, 1,1)
        }

        // Simple FPS limiting -- leads to microstutters without buffering, but totes fine for this
        now = time.Now().UnixNano()
        if now < nextTick {
            time.Sleep(time.Duration(nextTick - now))
        }
    }
}



func drawWorld(world *World, left,top int) {
    term.MoveCursor(left,top)
    width  := world.width
    height := world.height
    live   := 0
    var line string

    for y:=0; y<height; y++ {
        line = ""
        for x:=0; x<width; x++ {
            index := y*width + x
            if (*world.cells)[index] {
                neighbors := liveNeighborCount(world, x,y)
                line += colorByNeighbors(neighbors, "░")  // ░🮘
                // line += colorByNeighbors(neighbors, fmt.Sprintf("%v", neighbors))  // Displays neighbor count
                live++
            } else {
                line += " "
            }
        }
        term.MoveCursor(left, top+y)
        term.Printf(line)
    }
    // n.b. This draws some extra spaces to erase any artifacts from previous frames
    term.Printf("Size: %vx%v (%v)    Tick: %v        Live cells: %v    ", world.width,  world.height,  world.width*world.height,  world.tick,  live)
    term.Flush()
}



func tickWorld(world *World) *World {
    future := newWorld(world.width, world.height)
    future.tick = world.tick+1

    for x:=0; x<world.width; x++ {
        for y:=0; y<world.height; y++ {
            self      := y*world.width + x
            neighbors := liveNeighborCount(world, x, y)
            if        neighbors <  2 { (*future.cells)[self] = false                 // die
            } else if neighbors == 3 { (*future.cells)[self] = true                  // spawn
            } else if neighbors <  4 { (*future.cells)[self] = (*world.cells)[self]  // live
        //  } else if neighbors >  6 { (*future.cells)[self] = (*world.cells)[self]  // live -- for longevity
            } else                   { (*future.cells)[self] = false }               // die
        }
    }
    return future
}


func liveNeighborCount(world *World, x,y int) int {
    var neighbors int = 0
    // I'll need to rewrite this to allow non-wrapping
    if (*world.cells)[xyToCellIndex(world, x-1, y-1)] { neighbors++ }
    if (*world.cells)[xyToCellIndex(world, x,   y-1)] { neighbors++ }
    if (*world.cells)[xyToCellIndex(world, x+1, y-1)] { neighbors++ }
    if (*world.cells)[xyToCellIndex(world, x-1, y  )] { neighbors++ }
 // if (*world.cells)[xyToCellIndex(world, x,   y  )] { neighbors++ }
    if (*world.cells)[xyToCellIndex(world, x+1, y  )] { neighbors++ }
    if (*world.cells)[xyToCellIndex(world, x-1, y+1)] { neighbors++ }
    if (*world.cells)[xyToCellIndex(world, x,   y+1)] { neighbors++ }
    if (*world.cells)[xyToCellIndex(world, x+1, y+1)] { neighbors++ }
    return neighbors
}


func xyToCellIndex(world *World, x,y int) int {
    if x<0             { x += world.width  } // Wrap-around because it's more interesting
    if y<0             { y += world.height } //  .
    if x>=world.width  { x -= world.width  } //  .
    if y>=world.height { y -= world.height } //  .
    return y * world.width + x
}


func colorByNeighbors(neighbors int, text string) string {
    var color string
    reset := "\033[0;39m" + "\033[0;49m"
    if neighbors == 0        { color = "\033[0;30;44m"; /* black on blue         */
    } else if neighbors == 1 { color = "\033[1;30;46m"; /* black on bold cyan    */
    } else if neighbors == 2 { color = "\033[0;30;46m"; /* black on cyan         */
    } else if neighbors == 3 { color = "\033[0;30;42m"; /* black on green        */
    } else if neighbors == 4 { color = "\033[1;30;42m"; /* black on bold green   */
    } else if neighbors == 5 { color = "\033[0;30;41m"; /* black on red          */
    } else if neighbors == 6 { color = "\033[0;30;41m"; /* black on red          */
    } else if neighbors == 7 { color = "\033[1;30;41m"; /* black on bold red     */
    } else if neighbors == 8 { color = "\033[1;37;45m"; /* black on bold magenta */
    }
    return color + text + reset
}
