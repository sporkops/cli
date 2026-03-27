package output

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// GraphPoint represents a single data point for the response time graph.
type GraphPoint struct {
	ResponseTimeMs int64
	StatusMatch    bool   // true if status code matched expected
	Label          string // time label for X-axis
}

// PrintResponseGraph renders an ASCII scatter plot of response times.
// Points are colored green when the status matched and red when it didn't.
func PrintResponseGraph(points []GraphPoint) {
	if len(points) == 0 {
		return
	}

	// Skip graph if not a terminal (piped output)
	if !colorEnabled() {
		return
	}

	// Reverse so oldest is on the left, newest on the right
	reversed := make([]GraphPoint, len(points))
	for i, p := range points {
		reversed[len(points)-1-i] = p
	}
	points = reversed

	// Terminal width
	plotWidth := 60
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 20 {
		plotWidth = w - 12 // reserve space for Y-axis labels
		if plotWidth > 120 {
			plotWidth = 120
		}
	}

	const graphHeight = 10

	// Find min/max response times
	minY := points[0].ResponseTimeMs
	maxY := points[0].ResponseTimeMs
	for _, p := range points {
		if p.ResponseTimeMs < minY {
			minY = p.ResponseTimeMs
		}
		if p.ResponseTimeMs > maxY {
			maxY = p.ResponseTimeMs
		}
	}

	// Add padding so points aren't on the edges
	if maxY == minY {
		// All same value — create a range around it
		padding := maxY / 4
		if padding < 10 {
			padding = 10
		}
		minY -= padding
		maxY += padding
	} else {
		rangeY := maxY - minY
		maxY += rangeY / 10
		minY -= rangeY / 10
	}
	if minY < 0 {
		minY = 0
	}

	// Build the grid: grid[row][col] — row 0 is the top (maxY)
	type cell struct {
		hasPoint    bool
		statusMatch bool
	}
	grid := make([][]cell, graphHeight)
	for r := range grid {
		grid[r] = make([]cell, plotWidth)
	}

	// Place each point on the grid
	for i, p := range points {
		var col int
		if len(points) == 1 {
			col = plotWidth / 2
		} else {
			col = i * (plotWidth - 1) / (len(points) - 1)
		}
		if col >= plotWidth {
			col = plotWidth - 1
		}

		var row int
		yRange := maxY - minY
		if yRange > 0 {
			normalized := float64(p.ResponseTimeMs-minY) / float64(yRange)
			row = graphHeight - 1 - int(normalized*float64(graphHeight-1)+0.5)
		} else {
			row = graphHeight / 2
		}
		if row < 0 {
			row = 0
		}
		if row >= graphHeight {
			row = graphHeight - 1
		}

		grid[row][col] = cell{hasPoint: true, statusMatch: p.StatusMatch}
	}

	// Render the graph
	fmt.Println()
	for r := 0; r < graphHeight; r++ {
		// Y-axis label: show at top, middle, bottom
		var label string
		switch r {
		case 0:
			label = fmt.Sprintf("%6dms", maxY)
		case graphHeight / 2:
			label = fmt.Sprintf("%6dms", (maxY+minY)/2)
		case graphHeight - 1:
			label = fmt.Sprintf("%6dms", minY)
		default:
			label = "        "
		}

		fmt.Printf("%s │", label)

		for c := 0; c < plotWidth; c++ {
			if grid[r][c].hasPoint {
				if grid[r][c].statusMatch {
					fmt.Print(colorGreen + "●" + colorReset)
				} else {
					fmt.Print(colorRed + "●" + colorReset)
				}
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}

	// X-axis line
	fmt.Printf("         └%s", strings.Repeat("─", plotWidth))
	fmt.Println()

	// X-axis labels: show first, middle, and last timestamps
	if len(points) >= 2 {
		first := points[0].Label
		last := points[len(points)-1].Label
		mid := points[len(points)/2].Label

		spacing := plotWidth - len(first) - len(mid) - len(last)
		if spacing < 2 {
			// Just show first and last
			spacing = plotWidth - len(first) - len(last)
			if spacing < 1 {
				spacing = 1
			}
			fmt.Printf("          %s%s%s", first, strings.Repeat(" ", spacing), last)
		} else {
			leftGap := spacing / 2
			rightGap := spacing - leftGap
			fmt.Printf("          %s%s%s%s%s", first, strings.Repeat(" ", leftGap), mid, strings.Repeat(" ", rightGap), last)
		}
		fmt.Println()
	} else if len(points) == 1 {
		padding := plotWidth/2 - len(points[0].Label)/2
		if padding < 0 {
			padding = 0
		}
		fmt.Printf("          %s%s", strings.Repeat(" ", padding), points[0].Label)
		fmt.Println()
	}

	fmt.Println()
}
