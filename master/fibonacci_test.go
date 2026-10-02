package master

import (
	"env/cli"
	"env/cli/editor"
	"env/cli/tabs"
	"env/settings"
	"env/types"
	"testing"
)

func TestSelectReservesMonocleTabs(t *testing.T) {
	for _, mode := range []string{"monocle", "fibonacci"} {
		for _, size := range []types.Size{{100, 30}, {30, 100}, {1, 1}} {
			m := &Master{
				Mode: mode, Layout: "select",
				Sizes:   types.Dimensions{Full: size, Size: size},
				Clients: []cli.Client{&editor.Editor{Spec: "rec"}},
			}
			area := m.Sizes
			area.Size[1] = max(0, size[1]-3)
			area.Pos[1] = min(2, size[1])
			content := m.sidebarArea(area)
			wantPos := m.Sizes.Pos
			wantSize := types.Size{size[0], max(0, size[1]-1)}
			if mode == "monocle" {
				wantPos[1] += min(2, wantSize[1])
				wantSize[1] = max(0, size[1]-3)
			}
			if m.Sidebar.Sizes.Pos != wantPos || m.Sidebar.Sizes.Size != wantSize || content.Size != (types.Size{}) {
				t.Fatalf("%s %v: unexpected select geometry: %+v, content %+v", mode, size, m.Sidebar.Sizes, content)
			}
			m.Layout = "split"
			content = m.sidebarArea(area)
			if size[0] > 1 && (content.Size[0] == 0 || content.Size[1] == 0) {
				t.Fatalf("%s %v: split did not restore content space", mode, size)
			}
		}
	}
}

type fibonacciClient struct {
	cli.Client
	sizes types.Dimensions
}

func (c *fibonacciClient) Resize(sizes types.Dimensions) { c.sizes = sizes }

func TestFibonacciTabFocus(t *testing.T) {
	m := &Master{
		Mode:    "fibonacci",
		Sizes:   types.Dimensions{Full: types.Size{40, 15}, Size: types.Size{40, 15}},
		Clients: []cli.Client{&inputClient{}, &inputClient{}},
		Settings: &settings.Settings{Theme: types.Theme{
			Background: types.Color{R: 10}, Foreground: types.Color{R: 200},
		}},
	}
	m.Tabs = tabs.CreateTabs(m)
	panes, _ := m.fibonacciLayout()
	for _, focus := range []int{0, 1, 0} {
		m.SetFocus(focus)
		queues := m.Tabs.DrawFibonacci(panes)
		if len(queues) != 2 || len(m.Tabs.Draw().Frames) != 0 {
			t.Fatal("Fibonacci should show individual tabs without the global bar")
		}
		for i, q := range queues {
			frame := q.Frames[0]
			if frame.Sizes.Pos != panes[i].Pos || frame.Sizes.Size != (types.Size{panes[i].Size[0], 1}) {
				t.Fatal("tab must occupy only the first row of its own pane")
			}
			color := 'X'
			if i == focus {
				color = 'Y'
			}
			for x := 0; x < frame.Sizes.Size[0]; x++ {
				cell := frame.Cells[frame.Sizes.Pos[1]*40+frame.Sizes.Pos[0]+x]
				if cell.Ansi.Bg != m.Settings.ChooseColor(color, "bg", 0.5) || cell.Bold != (i == focus) {
					t.Fatalf("tab %d did not reflect focus %d across its entire width", i, focus)
				}
			}
		}
	}
}

func TestFibonacciSpiral(t *testing.T) {
	m := &Master{Mode: "fibonacci", Sizes: types.Dimensions{
		Full: types.Size{80, 43}, Size: types.Size{80, 43},
	}}
	want := []struct{ pos, size [2]int }{
		{[2]int{0, 1}, [2]int{39, 41}},
		{[2]int{40, 1}, [2]int{40, 19}},
		{[2]int{61, 22}, [2]int{19, 20}},
		{[2]int{40, 33}, [2]int{20, 9}},
		{[2]int{40, 22}, [2]int{9, 9}},
		{[2]int{50, 22}, [2]int{10, 9}},
	}
	for range want {
		m.Clients = append(m.Clients, &fibonacciClient{})
	}
	m.resizeFibonacci()
	for i, client := range m.Clients {
		got := client.(*fibonacciClient).sizes
		if got.Pos != types.Pos(want[i].pos) || got.Size != types.Size(want[i].size) || got.Full != m.Sizes.Full {
			t.Fatalf("pane %d: got %+v, want %+v", i, got, want[i])
		}
	}
	first := m.Clients[0].(*fibonacciClient)
	m.MoveClient(0, 2)
	if first.sizes.Pos != types.Pos(want[2].pos) || first.sizes.Size != types.Size(want[2].size) {
		t.Fatalf("reordered pane kept its old geometry: %+v", first.sizes)
	}
}

func TestFibonacciCoverage(t *testing.T) {
	for _, size := range []types.Size{{81, 44}, {1, 4}, {0, 0}, {2, 2}} {
		for count := 0; count <= 16; count++ {
			m := &Master{Sizes: types.Dimensions{Full: types.Size{100, 60}, Pos: types.Pos{3, 5}, Size: size}}
			for range count {
				m.Clients = append(m.Clients, &fibonacciClient{})
			}
			m.resizeFibonacci()
			panes, lines := m.fibonacciLayout()
			seen := map[types.Pos]bool{}
			for pos := range lines {
				seen[pos] = true
			}
			for i, client := range m.Clients {
				pane := panes[i]
				content := client.(*fibonacciClient).sizes
				if content.Pos[1] != pane.Pos[1]+min(1, pane.Size[1]) || content.Size[1] != max(0, pane.Size[1]-1) {
					t.Fatal("content must reserve one row for its tab")
				}
				if pane.Size[0] < 0 || pane.Size[1] < 0 {
					t.Fatalf("negative size: %+v", pane)
				}
				for y := pane.Pos[1]; y < pane.Pos[1]+pane.Size[1]; y++ {
					for x := pane.Pos[0]; x < pane.Pos[0]+pane.Size[0]; x++ {
						pos := types.Pos{x, y}
						if x < 3 || x >= 3+size[0] || y < 5 || y >= 5+size[1]-1 || seen[pos] {
							t.Fatalf("size %v, count %d: overlapping or out-of-bounds cell %v", size, count, pos)
						}
						seen[pos] = true
					}
				}
			}
			if count > 0 && len(seen) != size[0]*max(0, size[1]-1) {
				t.Fatalf("size %v, count %d: area left uncovered", size, count)
			}
		}
	}
}
