// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"cmp"
	"math"
	"slices"
	"strconv"

	"cogentcore.org/core/math32"
	"cogentcore.org/lab/plot"
)

// contourGrid computes the contour lines and filled regions of a grid of
// values. Each cell with four values is split into four triangles around
// its center, which has the mean of the four values, and the values are
// interpolated linearly within each triangle. This avoids the ambiguous
// saddle cells of marching squares, and gives lines and filled regions
// that match exactly, as both use the same points on the edges of the
// triangles, computed in the same way for the triangles sharing an edge.
type contourGrid struct {
	// rows, cols are the number of rows and columns of the grid.
	rows, cols int

	// x, y are the coordinates of the columns and rows.
	x, y []float64

	// vx, vy, vz are the coordinates and values of the vertices:
	// the grid points, row by row, followed by the cell centers.
	vx, vy, vz []float64

	// tris are the vertex indexes of the triangles, counter clockwise.
	tris [][3]int
}

// newContourGrid returns the contour grid of the given values, row by row,
// with the given coordinates of the columns and rows.
func newContourGrid(z []float64, rows, cols int, x, y []float64) *contourGrid {
	g := &contourGrid{rows: rows, cols: cols, x: x, y: y}
	n := rows * cols
	g.vx = make([]float64, n, n+(rows-1)*(cols-1))
	g.vy = make([]float64, n, cap(g.vx))
	g.vz = make([]float64, n, cap(g.vx))
	for i := range n {
		g.vx[i], g.vy[i], g.vz[i] = x[i%cols], y[i/cols], z[i]
	}
	for r := range rows - 1 {
		for c := range cols - 1 {
			a, b := r*cols+c, r*cols+c+1
			d, e := b+cols, a+cols
			za, zb, zd, ze := z[a], z[b], z[d], z[e]
			if math.IsNaN(za) || math.IsNaN(zb) || math.IsNaN(zd) || math.IsNaN(ze) {
				continue
			}
			ct := len(g.vx)
			g.vx = append(g.vx, 0.5*(x[c]+x[c+1]))
			g.vy = append(g.vy, 0.5*(y[r]+y[r+1]))
			g.vz = append(g.vz, 0.25*(za+zb+zd+ze))
			// a, b, d, e go around the cell counter clockwise, for
			// increasing coordinates, and the center is to the left
			g.tris = append(g.tris, [3]int{a, b, ct}, [3]int{b, d, ct}, [3]int{d, e, ct}, [3]int{e, a, ct})
		}
	}
	return g
}

// edge is an edge between two vertexes, lower index first.
type edge [2]int

// newEdge returns the edge between the given vertexes.
func newEdge(a, b int) edge {
	if a > b {
		a, b = b, a
	}
	return edge{a, b}
}

// point returns the point on the given edge at the given level, which is
// computed in the same way for each triangle with the edge.
func (g *contourGrid) point(e edge, level float64) math32.Vector2 {
	a, b := e[0], e[1]
	t := (level - g.vz[a]) / (g.vz[b] - g.vz[a])
	return math32.Vec2(float32(g.vx[a]+t*(g.vx[b]-g.vx[a])), float32(g.vy[a]+t*(g.vy[b]-g.vy[a])))
}

// vertex returns the position of the given vertex.
func (g *contourGrid) vertex(i int) math32.Vector2 {
	return math32.Vec2(float32(g.vx[i]), float32(g.vy[i]))
}

// lines returns the contour lines at the given level, as polylines,
// which are closed (the first point is repeated at the end) for loops.
func (g *contourGrid) lines(level float64) [][]math32.Vector2 {
	// segments between the edges crossing the level, in each triangle
	var segs [][2]edge
	for _, t := range g.tris {
		var cross [2]edge
		n := 0
		for i := range 3 {
			a, b := t[i], t[(i+1)%3]
			if (g.vz[a] >= level) != (g.vz[b] >= level) {
				cross[n] = newEdge(a, b)
				n++
			}
		}
		if n == 2 {
			segs = append(segs, cross)
		}
	}
	// chain the segments that share an edge into polylines
	ends := map[edge][]int{}
	for si, s := range segs {
		ends[s[0]] = append(ends[s[0]], si)
		ends[s[1]] = append(ends[s[1]], si)
	}
	used := make([]bool, len(segs))
	follow := func(start edge, si int) []edge {
		path := []edge{start}
		e := start
		for si >= 0 && !used[si] {
			used[si] = true
			s := segs[si]
			if s[0] == e {
				e = s[1]
			} else {
				e = s[0]
			}
			path = append(path, e)
			si = -1
			for _, o := range ends[e] {
				if !used[o] {
					si = o
					break
				}
			}
		}
		return path
	}
	var paths [][]edge
	// open lines start at an edge with one segment (at the boundary)
	for si, s := range segs {
		for _, e := range s {
			if !used[si] && len(ends[e]) == 1 {
				paths = append(paths, follow(e, si))
			}
		}
	}
	for si, s := range segs { // closed loops
		if !used[si] {
			paths = append(paths, follow(s[0], si))
		}
	}
	lines := make([][]math32.Vector2, len(paths))
	for i, p := range paths {
		pts := make([]math32.Vector2, len(p))
		for j, e := range p {
			pts[j] = g.point(e, level)
		}
		lines[i] = pts
	}
	return lines
}

// region returns the polygons of the region where the values are at or
// above the given level, one for each triangle with any of it, counter
// clockwise, so that filling all of them as one path has no seams.
func (g *contourGrid) region(level float64) [][]math32.Vector2 {
	var polys [][]math32.Vector2
	for _, t := range g.tris {
		var pts []math32.Vector2
		for i := range 3 {
			a, b := t[i], t[(i+1)%3]
			ina, inb := g.vz[a] >= level, g.vz[b] >= level
			if ina {
				pts = append(pts, g.vertex(a))
			}
			if ina != inb {
				pts = append(pts, g.point(newEdge(a, b), level))
			}
		}
		if len(pts) >= 3 {
			polys = append(polys, pts)
		}
	}
	return polys
}

// contourLevels returns the levels of the contours for the values with
// the given range, with their labels: the given levels within the range,
// if any, or else about n round values within the range, as for the
// ticks of an axis.
func contourLevels(mn, mx float64, n int, levels []float64) []plot.Tick {
	var lv []plot.Tick
	if len(levels) > 0 {
		vals := slices.Clone(levels)
		slices.Sort(vals)
		for _, l := range slices.Compact(vals) {
			if l >= mn && l <= mx {
				lv = append(lv, plot.Tick{Value: l, Label: strconv.FormatFloat(l, 'g', 4, 64)})
			}
		}
		return lv
	}
	if n <= 0 {
		n = 10
	}
	if mx <= mn {
		return nil
	}
	for _, t := range (plot.DefaultTicks{}).Ticks(mn, mx, n) {
		if t.IsMinor() || t.Value <= mn || t.Value > mx {
			continue
		}
		lv = append(lv, t)
	}
	slices.SortFunc(lv, func(a, b plot.Tick) int { return cmp.Compare(a.Value, b.Value) })
	return lv
}
