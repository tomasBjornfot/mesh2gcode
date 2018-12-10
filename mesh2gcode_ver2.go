package main

import (
	"os"
	"log"
	"encoding/json"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"strconv"
	"bytes"
)
/*
 * STRUCTS 
 */
// Strukt för hantering av settings från settings.json filen
type Settings struct {
	ToolRadius          float64    `json:"ToolRadius"`
	Xres            float64    `json:"Xres"`
	Yres            float64    `json:"Yres"`
}
/*
 * STUCTS
 */
type Mesh struct {
	Triangles [][9]float64
	Normals   [][3]float64
    Profile   [][2]float64
	No_tri    int
	X_min     float64
	Y_min     float64
	Z_min     float64
	X_max     float64
	Y_max     float64
	Z_max     float64
}
type CrossSection struct {
	X       [][1000]float64
	Y       [][1000]float64
	Z       [][1000]float64
	No_rows int
	No_cols [1000]int
}

type Mill struct {
	X       [][1000]float64
	Y       [][1000]float64
	Z       [][1000]float64
	Xn       [][1000]float64
	Yn       [][1000]float64
	Zn       [][1000]float64
	No_rows int
	No_cols [1000]int
}
/*
 * PRIVATE FUNCTIONS/METHODS
 */
func (mesh *Mesh) calculateMeshProperties() {
	// tar reda på max och min i varje dimension
	mesh.X_min, mesh.Y_min, mesh.Z_min = 100000.0, 100000.0, 100000.0
	mesh.X_max, mesh.Y_max, mesh.Z_max = -100000.0, -100000.0, -100000.0
	points := trianglesToPoints(*mesh)

	for i := 0; i < 3*mesh.No_tri; i++ {
		// min x
		if points[i][0] < mesh.X_min {
			mesh.X_min = points[i][0]
		}
		// min y
		if points[i][1] < mesh.Y_min {
			mesh.Y_min = points[i][1]
		}
		// min z
		if points[i][2] < mesh.Z_min {
			mesh.Z_min = points[i][2]
		}
		// max x
		if points[i][0] > mesh.X_max {
			mesh.X_max = points[i][0]
		}
		// max y
		if points[i][1] > mesh.Y_max {
			mesh.Y_max = points[i][1]
		}
		// max z
		if points[i][2] > mesh.Z_max {
			mesh.Z_max = points[i][2]
		}
	}
}
func (mesh *Mesh) calculateProfile(radius float64, resolution int) {
	// räknar ut Profilen på brädan i xy planet

	// plockar ut punkterna från trianglarna
	points := trianglesToPoints(*mesh)

	x := make([]float64, len(points))
	y := make([]float64, len(points))
	for i := 0; i < len(points); i++ {
		x[i] = points[i][1]
		y[i] = points[i][0]
	}

	// tar reda på y värden på Profilen genom att:
	// 0. skapa x värden som cikeln ska ramla ner på
	// 1. iterera över alla x drop värden
	// 2. iterera över alla punkter
	// 3. väljer bara punkter där y > 0
	// 4. tar ut dom punkter som ligger inom [x - radius,x + radius]
	// 5. beräkna höjd på cirkel som krockar med värdet
	// 6. ta ut index på punkten (som cirkeln krockar med)
	// 7. om det finns ett större värde, gäller detta

	// 0
	mesh.calculateMeshProperties()
	drop_x := linspace(mesh.Y_min+0.1-radius, mesh.Y_max-0.1+radius, resolution)
	// 1
	r2 := radius * radius
	index := make([]int, len(drop_x))
	for i := 0; i < len(drop_x); i++ {
		ymax := float64(-1)
		index[i] = -1
		// 2
		for j := 0; j < len(x); j++ {
			// 3
			if y[j] > 0 {
				// 4
				if x[j] > drop_x[i]-radius && x[j] < drop_x[i]+radius {
					// 5
					ymax_new := y[j] + math.Sqrt(r2+(x[j]-drop_x[i])*(x[j]-drop_x[i]))
					if ymax_new > ymax {
						// 6
						ymax = ymax_new
						index[i] = j
					}
				}
			}
		}
	}
	no_index := int(0)
	for i := 0; i < len(index); i++ {
		if index[i] != -1 {
			no_index++
		}
	}	
	mesh.Profile = make([][2]float64, no_index)
	pindex := int(0)
	for i := 0; i < len(index); i++ {
		if index[i] != -1 {
			mesh.Profile[pindex][0] = x[index[i]]
			mesh.Profile[pindex][1] = y[index[i]]
			pindex++
		}
	}
}
func (mesh *Mesh) calculateNormals() {
	// räknar ut normaler för trianglar
	var v0, v1 [3]float64
	for i := 0; i < mesh.No_tri; i++ {
		v0[0] = mesh.Triangles[i][0] - mesh.Triangles[i][6]
		v0[1] = mesh.Triangles[i][1] - mesh.Triangles[i][7]
		v0[2] = mesh.Triangles[i][2] - mesh.Triangles[i][8]

		v1[0] = mesh.Triangles[i][0] - mesh.Triangles[i][3]
		v1[1] = mesh.Triangles[i][1] - mesh.Triangles[i][4]
		v1[2] = mesh.Triangles[i][2] - mesh.Triangles[i][5]

		mesh.Normals[i] = crossProduct(v0, v1)
	}
}
func crossProduct(v0 [3]float64, v1 [3]float64) [3]float64 {
	var x [3]float64
	x[0] = v0[1]*v1[2] - v0[2]*v1[1]
	x[1] = v0[2]*v1[0] - v0[0]*v1[2]
	x[2] = v0[0]*v1[1] - v0[1]*v1[0]
	length := math.Sqrt(x[0]*x[0] + x[1]*x[1] + x[2]*x[2])
	x[0] = x[0] / length
	x[1] = x[1] / length
	x[2] = x[2] / length
	return x
}
func linspace(min float64, max float64, no_segments int) []float64 {
	numbers := make([]float64, no_segments+1)
	numbers[0] = min
	segment := (max - min) / float64(no_segments)
	for i := 1; i < len(numbers); i++ {
		numbers[i] = numbers[i-1] + segment
	}
	return numbers
}
func trianglesToPoints(mesh Mesh) [][3]float64 {
	// tar ut trianglarna från mesh och gör en x,3 matris av punkterna
	points := make([][3]float64, 3*mesh.No_tri)
	// plockar ut punkterna från trianglarna
	for i := 0; i < mesh.No_tri; i++ {
		points[3*i+0][0] = mesh.Triangles[i][0]
		points[3*i+1][0] = mesh.Triangles[i][3]
		points[3*i+2][0] = mesh.Triangles[i][6]

		points[3*i+0][1] = mesh.Triangles[i][1]
		points[3*i+1][1] = mesh.Triangles[i][4]
		points[3*i+2][1] = mesh.Triangles[i][7]

		points[3*i+0][2] = mesh.Triangles[i][2]
		points[3*i+1][2] = mesh.Triangles[i][5]
		points[3*i+2][2] = mesh.Triangles[i][8]
	}
	return points
}
func getMinValue(array []float64) int {
	min_value := float64(1000000)
	index := int(-1)
	for i := 0; i < len(array); i++ {
		if array[i] < min_value {
			min_value = array[i]
			index = i
		}
	}
	return index
}
func getMaxValue(array []float64) int {
	max_value := float64(-1000000)
	index := int(-1)
	for i := 0; i < len(array); i++ {
		if array[i] > max_value {
			max_value = array[i]
			index = i
		}
	}
	return index
}
func getNearestNeighbours(x float64, x_array []float64) []int {
	lower_diff := float64(100000)
	upper_diff := float64(100000)
	lower_index := int(-1)
	upper_index := int(-1)
	for i := 0; i < len(x_array); i++ {
		// prospekt för nedre värdet
		if x_array[i] < x {
			lower_diff_new := x - x_array[i]
			if lower_diff_new < lower_diff {
				lower_diff = lower_diff_new
				lower_index = i
			}
		}
		// prospekt för övre värdet
		if x_array[i] >= x {
			upper_diff_new := x_array[i] - x
			if upper_diff_new < upper_diff {
				upper_diff = upper_diff_new
				upper_index = i
			}
		}
	}
	index := make([]int, 2)
	index[0] = lower_index
	index[1] = upper_index
	return index
}
func twoPointsToLine(x0 float64, x1 float64, y0 float64, y1 float64) (float64, float64) {
	k := (y1 - y0) / (x1 - x0)
	m := y0 - k*x0
	return k, m
}
func yValueAt(x float64, k float64, m float64) float64 {
	return k*x + m
}
func unique(x [1000]float64, y [1000]float64, test_length int) ([1000]float64, [1000]float64, int){
	// makes unique points with respect to the x value
	// has a resolution of 2 decimals
	remove := make([]int, 1000)
	no_index := 0
	for i:=0; i<test_length-1; i++ {
		cx := x[i]
		for j:=i+1; j<test_length; j++ {	
			dx := cx - x[j]
			if math.Round(100*dx) == 0 {
				//fmt.Println("i:",i,"j:",j)
				remove[no_index] = j
				no_index++
				break
			}
		}
	}
	u_x := make([]float64, 1000)
	u_y := make([]float64, 1000)
	copy(u_x, x[:])
	copy(u_y, y[:])
	for i:=no_index-1; i>-1; i-- {
		u_x = append(u_x[:remove[i]], u_x[remove[i]+1:]...)
		u_y = append(u_y[:remove[i]], u_y[remove[i]+1:]...)
	}	

	var x_out [1000]float64
	copy(x_out[:], u_x[:])
	
	var y_out [1000]float64
	copy(y_out[:], u_y[:])
	//fmt.Println("remove:",remove[:no_index])
	//fmt.Println("test_length", test_length)
	//fmt.Println("no_index", no_index)
	return x_out, y_out, test_length - no_index 
}
func sort_index(arr [1000]float64, no_cols int) []int {
	get_min_index := func(arr []float64) int {
		val := 100000.0
		min_index := -1
		for i:=0; i<len(arr); i++ {
			if arr[i] < val {
				val = arr[i]
				min_index = i
			}
		}
		return min_index
	}
	
	index := make([]int, no_cols)
	arr_copy := make([]float64, no_cols)
	copy(arr_copy, arr[:no_cols])
	for i:=0; i<len(arr_copy); i++{
		index[i] = get_min_index(arr_copy)
		arr_copy[index[i]] = 1000000.0
	}
	return index
}
func sort2darray(x [1000]float64, y [1000]float64, no_cols int) ([1000]float64, [1000]float64) {
	index := sort_index(x, no_cols)
	var x_out [1000]float64
	var y_out [1000]float64
	for i:=0; i<no_cols; i++ {
		x_out[i] = x[index[i]]
		y_out[i] = y[index[i]]
	}
	return x_out, y_out
}
// --- 2D line calculations ---
func line_square_length(p0 [2]float64, p1 [2]float64) float64 {
	dx := p0[0] - p1[0]
	dy := p0[1] - p1[1]
	return dx*dx + dy*dy
}
func line_k(p0 [2]float64, p1 [2]float64) float64 {
	dx := p1[0] - p0[0]
	dy := p1[1] - p0[1]
	k := dy/dx
	if dx +dy < 0.1 {
		return 0
	}
	return k
}
func line_m(p0 [2]float64, p1 [2]float64) float64 {
	k := line_k(p0, p1)
	return p0[1] - k*p0[0]
}
func calc_next(p_left [2]float64, r2 float64, k float64, m float64) [2]float64 {
	dx := math.Sqrt(r2/(1+k*k))
	p_new := [2]float64 {0.0, 0.0}
	p_new[0] = p_left[0] + dx
	p_new[1] = k*p_new[0] + m
	return p_new
}
func even_spaced(x [1000]float64, y [1000]float64, no_cols int, space float64) ([]float64, []float64) {
	d2 := make([]float64, no_cols)
	k := make([]float64, no_cols)
	m := make([]float64, no_cols)
	d2_left := 0.0
	var p0, p1, p_start, p_end, p_last, new_point [2]float64
	for i:=0; i<no_cols; i++ {
		p0[0] = x[i] 
		p0[1] = y[i]
		p1[0] = x[i+1]
		p1[1] = y[i+1]
		d2[i] = line_square_length(p0, p1)
		k[i] = line_k(p0, p1)
		m[i] = line_m(p0, p1)	
	}
	//
	line_index := 0
	point_index := 0
	cs_x := make([]float64, 1000)
	cs_y := make([]float64, 1000)
	// take the first point in cs as a starting point
	cs_x[point_index] = x[0]
	cs_y[point_index] = y[0]
	for i:=0; i<100; i++ {
		//fmt.Println("i:",i)
		//check if new point pass the center line (x=0)
		if cs_x[point_index] > 0 {
			cs_x[point_index] = 0
			//cs_y[point_index] =  m[line_index]
			cs_y[point_index] =  cs_y[point_index - 1] 
			break
		}
		space2 := space*space
		// calculating the square distance from the previous point ...
		// to the end of the line that it belongs to (d2_left)
		p_end[0] = x[line_index+1]
		p_end[1] = y[line_index+1]
		p_last[0] = cs_x[point_index]
		p_last[1] = cs_y[point_index]
		d2_left = line_square_length(p_last, p_end)
		if space2 < d2_left {
			//... then the next point should be in this line
			// finds the point where the distance is space2 
			// from the last point to the new point
			p_start[0] = cs_x[point_index]
			p_start[1] = cs_y[point_index]
			new_point = calc_next(p_start, space2, k[line_index], m[line_index])
			point_index++
			cs_x[point_index] = new_point[0]
			cs_y[point_index] = new_point[1]
			/*
			fmt.Println("p_start", p_start)
			fmt.Println("space2", space2)
			fmt.Println("k[line_index]", k[line_index])
			fmt.Println("m[line_index]", m[line_index])
			fmt.Println("point index:", point_index)
			fmt.Println("cs_x[point_index]:",cs_x[point_index])
			fmt.Println("cs_y[point_index]:",cs_y[point_index])
			*/		
		} else {
			for {
				// new space2 when looking at the next line...
				space2 = (math.Sqrt(space2) - math.Sqrt(d2_left))*(math.Sqrt(space2) - math.Sqrt(d2_left))
				// move index to next line
				line_index = line_index + 1
				// move to next line distance. Note that d2_left is the whole
				// line length since previous point was on the previous line
				d2_left = d2[line_index]
				if space2 < d2_left {
					// ... then the next point should be in this line
					// finds the point where the distance is space2 
					p_start[0] = x[line_index]
					p_start[1] = y[line_index]
					
					new_point = calc_next(p_start, space2, k[line_index], m[line_index])
					point_index++
					cs_x[point_index] = new_point[0]
					cs_y[point_index] = new_point[1]
					/*
					fmt.Println("Next line")
					fmt.Println("p_start", p_start)
					fmt.Println("space2", space2)
					fmt.Println("k[line_index]", k[line_index])
					fmt.Println("m[line_index]", m[line_index])
					fmt.Println("point index:", point_index)
					fmt.Println("cs_x[point_index]:",cs_x[point_index])
					fmt.Println("cs_y[point_index]:",cs_y[point_index])
					*/
					break
				}
			}
		}
			
	}
	cs_x_out := cs_x[:(point_index+1)] 
	cs_y_out := cs_y[:(point_index+1)] 
	return cs_x_out, cs_y_out
}
/*
 * PUBLIC FUNCTIONS/METHODS
 */
func (mesh *Mesh) Read(path string, filetype int) {
	// läser in en STL fil till mesh
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("ReadFromFile: Something went wrong!!!")
	}
	defer file.Close()
	finfo, _ := file.Stat()

	bytes := make([]byte, finfo.Size())
	file.Read(bytes)

	var row int
	s_array := strings.Split(string(bytes), "\n")
	if filetype == 1 {
		mesh.No_tri = (len(s_array)-2)/7 - 1
		mesh.Triangles = make([][9]float64, mesh.No_tri)
		mesh.Normals = make([][3]float64, mesh.No_tri)
		for i := 1; i < mesh.No_tri+1; i++ {
			s_tri := s_array[7*i+1 : 7*(i+1)+1]
			n_string := strings.Split(s_tri[0], " ")

			// läser in normalerna till triangeln
			for j := 0; j < 3; j++ {
				mesh.Normals[row][j], _ = strconv.ParseFloat(n_string[j+2], 32)
			}
			// läser in hörnpunkterna till triangeln
			t1_string := strings.Split(s_tri[2], " ")
			t2_string := strings.Split(s_tri[4], " ")
			t3_string := strings.Split(s_tri[3], " ")
			for j := 0; j < 3; j++ {
				mesh.Triangles[row][j+0], _ = strconv.ParseFloat(t1_string[j+3], 32)
				mesh.Triangles[row][j+3], _ = strconv.ParseFloat(t2_string[j+3], 32)
				mesh.Triangles[row][j+6], _ = strconv.ParseFloat(t3_string[j+3], 32)
			}
			row++
		}
	}
	if filetype == 0 {
		mesh.No_tri = (len(bytes) - 84) / 50
		mesh.Triangles = make([][9]float64, mesh.No_tri)
		mesh.Normals = make([][3]float64, mesh.No_tri)
		bits := binary.LittleEndian.Uint32(bytes[0:4])
		var start_index int = 80 // hoppar över header
		start_index += 4         //  hoppar över uint32
		for i := 0; i < mesh.No_tri; i++ {
			// normaler
			for j := 0; j < 3; j++ {
				bits = binary.LittleEndian.Uint32(bytes[(start_index):(start_index + 4)])
				mesh.Normals[i][j] = float64(math.Float32frombits(bits))
				start_index += 4
			}
			// trianglar
			for j := 0; j < 9; j++ {
				bits = binary.LittleEndian.Uint32(bytes[(start_index):(start_index + 4)])
				mesh.Triangles[i][j] = float64(math.Float32frombits(bits))
				start_index += 4
			}
			// hoppar över uint16
			start_index += 2
		}
	}
	mesh.calculateProfile(50.0, 100)
}
func (mesh *Mesh) Write(path string) {
	// skriver en mesh till binär STL fil
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("WriteToFile: Something went wrong!!!")
	}
	defer file.Close()

	header := make([]byte, 80)
	file.Write(header)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, int32(mesh.No_tri))
	for i := 0; i < mesh.No_tri; i++ {
		for j := 0; j < 3; j++ {
			binary.Write(buf, binary.LittleEndian, float32(mesh.Normals[i][j]))
		}

		for j := 0; j < 9; j++ {
			binary.Write(buf, binary.LittleEndian, float32(mesh.Triangles[i][j]))
		}

		binary.Write(buf, binary.LittleEndian, int16(i))
	}
	buf.WriteTo(file)
}
func (mesh *Mesh) MoveToCenter() {
	// flyttar en mesh till centrum
	var x_sum, y_sum, z_sum float64 = 0, 0, 0
	var vector [3]float64
	for i := 0; i < mesh.No_tri; i++ {
		x_sum += mesh.Triangles[i][0]
		x_sum += mesh.Triangles[i][3]
		x_sum += mesh.Triangles[i][6]

		y_sum += mesh.Triangles[i][1]
		y_sum += mesh.Triangles[i][4]
		y_sum += mesh.Triangles[i][7]

		z_sum += mesh.Triangles[i][2]
		z_sum += mesh.Triangles[i][5]
		z_sum += mesh.Triangles[i][8]
	}
	vector[0] = -x_sum / (3 * float64(mesh.No_tri))
	vector[1] = -y_sum / (3 * float64(mesh.No_tri))
	vector[2] = -z_sum / (3 * float64(mesh.No_tri))
	fmt.Println("MovetoCenter, translation vector: ", vector)
	mesh.Translate(vector)
}
func (mesh *Mesh) MoveToCenter2() {
	// flyttar till centrum map högsta och minsta värde
	var X_min, Y_min, Z_min float64 = 1000.0, 1000.0, 1000.0
	var X_max, Y_max, Z_max float64 = -1000.0, -1000.0, -1000.0
	var vector [3]float64
	for i := 0; i < mesh.No_tri; i++ {
		for j := 0; j < 7; j = j + 3 {
			if mesh.Triangles[i][j] < X_min {
				X_min = mesh.Triangles[i][j]
			}
			if mesh.Triangles[i][j] > X_max {
				X_max = mesh.Triangles[i][j]
			}
		}
		for j := 1; j < 8; j = j + 3 {
			if mesh.Triangles[i][j] < Y_min {
				Y_min = mesh.Triangles[i][j]
			}
			if mesh.Triangles[i][j] > Y_max {
				Y_max = mesh.Triangles[i][j]
			}
		}
		for j := 2; j < 9; j = j + 3 {
			if mesh.Triangles[i][j] < Z_min {
				Z_min = mesh.Triangles[i][j]
			}
			if mesh.Triangles[i][j] > Z_max {
				Z_max = mesh.Triangles[i][j]
			}
		}
	}
	vector[0] = -(X_max + X_min) / 2.0
	vector[1] = -(Y_max + Y_min) / 2.0
	vector[2] = -(Z_max + Z_min) / 2.0
	fmt.Println("MovetoCenter2, translation vector: ", vector)
	mesh.Translate(vector)
}
func (mesh *Mesh) Translate(vector [3]float64) {
	// translaterar en mesh i rymden
	for i := 0; i < mesh.No_tri; i++ {
		for j := 0; j < 9; j = j + 3 {
			mesh.Triangles[i][j] = mesh.Triangles[i][j] + vector[0]
		}
		for j := 1; j < 9; j = j + 3 {
			mesh.Triangles[i][j] = mesh.Triangles[i][j] + vector[1]
		}
		for j := 2; j < 9; j = j + 3 {
			mesh.Triangles[i][j] = mesh.Triangles[i][j] + vector[2]
		}
	}
}
func (mesh *Mesh) Rotate(axis string, angle_degrees float64) {
	// roterar en mesh i rymden
	ar := angle_degrees * math.Pi / 180.0
	for i := 0; i < mesh.No_tri; i++ {
		for j := 0; j < 9; j = j + 3 {
			point := mesh.Triangles[i][j:(j + 3)]
			new_point := make([]float64, 3)
			if axis == "x" {
				new_point[0] = point[0]
				new_point[1] = point[1]*math.Cos(ar) - point[2]*math.Sin(ar)
				new_point[2] = point[1]*math.Sin(ar) + point[2]*math.Cos(ar)
			}
			if axis == "y" {
				new_point[0] = point[0]*math.Cos(ar) + point[2]*math.Sin(ar)
				new_point[1] = point[1]
				new_point[2] = -point[0]*math.Sin(ar) + point[2]*math.Cos(ar)
			}
			if axis == "z" {
				new_point[0] = point[0]*math.Cos(ar) - point[1]*math.Sin(ar)
				new_point[1] = point[0]*math.Sin(ar) + point[1]*math.Cos(ar)
				new_point[2] = point[2]
			}
			mesh.Triangles[i][j+0] = new_point[0]
			mesh.Triangles[i][j+1] = new_point[1]
			mesh.Triangles[i][j+2] = new_point[2]
		}
	}
	mesh.calculateNormals()
}
func (mesh *Mesh) AlignMesh(cadtype string) {
	// gör en alignment beroende på vilkan cad som används
	if cadtype == "boardcad" {
		mesh.MoveToCenter()
		mesh.Rotate("x", 90)
		mesh.Rotate("z", 90)
	}
	mesh.calculateNormals()
}
func (mesh *Mesh) AlignMeshX() {
	// roterar brädan runt x vectorn tills man hittar ett minimum Z_max - Z_min
	mesh.calculateMeshProperties()
	rmesh := mesh
	z_range := rmesh.Z_max - rmesh.Z_min
	for i := 0; i < 50; i++ {
		rmesh.Rotate("x", -0.1)
		rmesh.calculateMeshProperties()
		if rmesh.Z_max-rmesh.Z_min < z_range {
			z_range = rmesh.Z_max - rmesh.Z_min
		} else {
			//fmt.Printf("Alignment x rotation: %0.2f degrees\n", 0.1*float64(i))
			break
		}
	}
	mesh = rmesh
}
func (mesh *Mesh) Split() (*Mesh, *Mesh) {
	// delar upp deck och bottom på brädan
	// flytta funktionen
	No_tri_deck := int(0)
	No_tri_bottom := int(0)
	for i := 0; i < mesh.No_tri; i++ {
		if mesh.Normals[i][2] < 0 {
			No_tri_deck++
		}
		if mesh.Normals[i][2] >= 0 {
			No_tri_bottom++
		}
	}
	deck := new(Mesh)
	deck.Triangles = make([][9]float64, No_tri_deck)
	deck.Normals = make([][3]float64, No_tri_deck)
	deck.No_tri = No_tri_deck

	bottom := new(Mesh)
	bottom.Triangles = make([][9]float64, No_tri_bottom)
	bottom.Normals = make([][3]float64, No_tri_bottom)
	bottom.No_tri = No_tri_bottom

	i_deck := int(0)
	i_bottom := int(0)
	for i := 0; i < mesh.No_tri; i++ {
		if mesh.Normals[i][2] < 0 {
			deck.Triangles[i_deck] = mesh.Triangles[i]
			deck.Normals[i_deck] = mesh.Normals[i]
			i_deck++
		}
		if mesh.Normals[i][2] >= 0 {
			bottom.Triangles[i_bottom] = mesh.Triangles[i]
			bottom.Normals[i_bottom] = mesh.Normals[i]
			i_bottom++
		}
	}
	deck.calculateProfile(50.0, 100)
	bottom.calculateProfile(50.0, 100)
	return deck, bottom
}
func (mesh *Mesh) CalculateCS_Y_Values(max_distance float64, resolution float64) []float64 {
	// Räknar ut y värden som ska navändas som cross sections
	// hämtar profilen
	px := make([]float64, len(mesh.Profile))
	py := make([]float64, len(mesh.Profile))
	for i := 0; i < len(mesh.Profile); i++ {
		px[i] = mesh.Profile[i][0]
		py[i] = mesh.Profile[i][1]
	}
	// skapar "cross sections" cs_x och cs_y
	cs_x := make([]float64, 100000)
	cs_y := make([]float64, 100000)
	start_index := getMinValue(px)
	stop_index := getMaxValue(px)
	cs_index := int(0)
	nindex := make([]int, 2)
	// ger ett startvärde (minsta värdet i profilen)
	cs_x[0] = px[start_index]
	cs_y[0] = py[start_index]
	cs_x_new := cs_x[0]
	cs_y_new := float64(0)
	k := float64(0)
	m := float64(0)
	max_distance2 := max_distance * max_distance
	dist2 := float64(0)

	// ** iteration börjar **
	for i := 0; i < 100000; i++ {
		// flyttar sig frammåt med en resolution
		cs_x_new += resolution
		// kollar så cs_x_new inte är större än max värdet
		if cs_x_new == px[stop_index] {
			break
		}
		if cs_x_new > px[stop_index] {
			cs_index++
			cs_x[cs_index] = px[stop_index]
			break
		}
		// hittar närmsta grannar
		nindex = getNearestNeighbours(cs_x_new, px)
		// räknar ut y värdet
		k, m = twoPointsToLine(px[nindex[0]], px[nindex[1]], py[nindex[0]], py[nindex[1]])
		cs_y_new = yValueAt(cs_x_new, k, m)
		// räknar ut avståndet mellan cs punkterna
		dist2 = (cs_x_new-cs_x[cs_index])*(cs_x_new-cs_x[cs_index]) + (cs_y_new-cs_y[cs_index])*(cs_y_new-cs_y[cs_index])
		// kollar om dom nya punkterna har passerat max_distance
		if dist2 > max_distance2 {
			cs_x_new -= resolution
			cs_index++
			cs_x[cs_index] = cs_x_new
			cs_y[cs_index] = cs_y_new
		}
	}
	cs_x_final := make([]float64, cs_index+1)
	for i := 0; i < cs_index+1; i++ {
		cs_x_final[i] = cs_x[i]
	}
	return cs_x_final
}
func (crossSection *CrossSection) MeshToCs(cs []float64, mesh *Mesh) {
	// return variabler
	x := make([][1000]float64, len(cs))
	y := make([][1000]float64, len(cs))
	z := make([][1000]float64, len(cs))
	var no_cols [1000]int

	// variabler som används endast i funktionen
	var side [3]int
	var v0 [3]float64
	var p0 []float64
	var tri [9]float64
	var t float64
	var index, side_sum int = 0, 0

	// itererar över alla tvärsnitt
	for i := 0; i < len(cs); i++ {
		index = 0
		// itererar över all trianglar
		for j := 0; j < mesh.No_tri; j++ {
			if mesh.Triangles[j][1]-cs[i] > 0 {
				side[0] = 1
			} else {
				side[0] = -1
			}
			if mesh.Triangles[j][4]-cs[i] > 0 {
				side[1] = 1
			} else {
				side[1] = -1
			}
			if mesh.Triangles[j][7]-cs[i] > 0 {
				side[2] = 1
			} else {
				side[2] = -1
			}
			// om y korsar triangeln
			side_sum = side[0] + side[1] + side[2]
			if side_sum == 1 || side_sum == -1 {
				tri = mesh.Triangles[j]
				if side[0]+side[1] == 0 {
					v0[0] = tri[3] - tri[0]
					v0[1] = tri[4] - tri[1]
					v0[2] = tri[5] - tri[2]
					p0 = tri[0:3]
					t = (cs[i] - p0[1]) / v0[1]
					x[i][index] = v0[0]*t + p0[0]
					z[i][index] = v0[2]*t + p0[2]
					index++
				}
				if side[1]+side[2] == 0 {
					v0[0] = tri[6] - tri[3]
					v0[1] = tri[7] - tri[4]
					v0[2] = tri[8] - tri[5]
					p0 = tri[3:6]
					t = (cs[i] - p0[1]) / v0[1]
					x[i][index] = v0[0]*t + p0[0]
					z[i][index] = v0[2]*t + p0[2]
					index++
				}
				if side[0]+side[2] == 0 {
					v0[0] = tri[6] - tri[0]
					v0[1] = tri[7] - tri[1]
					v0[2] = tri[8] - tri[2]
					p0 = tri[0:3]
					t = (cs[i] - p0[1]) / v0[1]
					x[i][index] = v0[0]*t + p0[0]
					z[i][index] = v0[2]*t + p0[2]
					index++
				}
			}
		}
		no_cols[i] = index
		x[i], z[i], no_cols[i] = unique(x[i],z[i], index)
		x[i], z[i] = sort2darray(x[i], z[i], no_cols[i])
		
	}	
	crossSection.No_cols = no_cols
	crossSection.No_rows = len(cs)
	crossSection.X = x
	crossSection.Z = z

	for i := 0; i < len(cs); i++ {
		for j := 0; j < 1000; j++ {
			y[i][j] = cs[i]
		}
	}
	crossSection.Y = y
}
func (mesh *Mesh) CalculateCrossSections(y_res float64, m_res float64) *CrossSection {
	cs := mesh.CalculateCS_Y_Values(y_res, m_res)
	cs_mesh := new(CrossSection)
    cs_mesh.MeshToCs(cs, mesh)
    return cs_mesh
}
func (mill *Mill) CalculateMillNormals() {
	// bekrivning!
	nx := make([][1000]float64, mill.No_rows)
	ny := make([][1000]float64, mill.No_rows)
	nz := make([][1000]float64, mill.No_rows)
	var v0 [3]float64
	var v1 [3]float64
	var n [3]float64
	
	// räknar ut max no_cols
	max_no_cols := 0
	for nc := range mill.No_cols {
		if nc > max_no_cols {
			max_no_cols = nc
		}
	}
	
	// beräknar alla utom kanterna
	for i := 1; i < mill.No_rows-1; i++ {
		for j := 1; j < max_no_cols-1; j++ {
			v0[0] = mill.X[i][j+1] - mill.X[i][j-1]
			v0[1] = mill.Y[i][j+1] - mill.Y[i][j-1]
			v0[2] = mill.Z[i][j+1] - mill.Z[i][j-1]

			v1[0] = mill.X[i+1][j] - mill.X[i-1][j]
			v1[1] = mill.Y[i+1][j] - mill.Y[i-1][j]
			v1[2] = mill.Z[i+1][j] - mill.Z[i-1][j]
			n = crossProduct(v0, v1)
			l := math.Sqrt(n[0]*n[0] + n[1]*n[1] + n[2]*n[2])
			nx[i][j] = n[0] / l
			ny[i][j] = n[1] / l
			nz[i][j] = n[2] / l
		}
	}
	// första och sista kolumnen
	for i := 0; i < mill.No_rows; i++ {
		// tar grannens vaäde
		nx[i][0] = nx[i][1]
		ny[i][0] = ny[i][1]
		nz[i][0] = nz[i][1]

		nx[i][max_no_cols-1] = nx[i][max_no_cols-2]
		ny[i][max_no_cols-1] = ny[i][max_no_cols-2]
		nz[i][max_no_cols-1] = nz[i][max_no_cols-2]
	}
	// första och sista raden
	for i := 0; i < max_no_cols; i++ {
		// tar grannens värde
		nx[0][i] = nx[1][i]
		ny[0][i] = ny[1][i]
		nz[0][i] = nz[1][i]

		nx[mill.No_rows-1][i] = nx[mill.No_rows-2][i]
		ny[mill.No_rows-1][i] = ny[mill.No_rows-2][i]
		nz[mill.No_rows-1][i] = nz[mill.No_rows-2][i]
	}

	mill.Xn = nx
	mill.Yn = ny
	mill.Zn = nz
}
func (mill *Mill) CalculateMillCoordinates(cs *CrossSection, settings *Settings) {
	
	mill.X = make([][1000]float64, cs.No_rows)
	mill.Y = make([][1000]float64, cs.No_rows)
	mill.Z = make([][1000]float64, cs.No_rows)
	
	//fmt.Println("cs.X:",cs.X[1])
	//fmt.Println("cs.Z:",cs.Z[1])
	//even_spaced(cs.X[1], cs.Z[1], cs.No_cols[1], 4.0)
	
	mill.No_rows = cs.No_rows
	for i:=0; i<cs.No_rows; i++ {
		cs_x, cs_z := even_spaced(cs.X[i], cs.Z[i], cs.No_cols[i], settings.Xres)
		for j:=0; j<1000; j++ {
			mill.Y[i][j] = cs.Y[i][0]
			if j<len(cs_x) {
				mill.X[i][j] = cs_x[j]
				mill.Z[i][j] = cs_z[j]
			} else {
				mill.X[i][j] = cs_x[len(cs_x)-1]
				mill.Z[i][j] = cs_z[len(cs_x)-1]
			}
		}
		mill.No_cols[i] = len(cs_x)
	}	
}
/*
 * EXTRAS
 */
func (mesh *Mesh) WritePointsToFile(path string) {
	// skriver alla triangelhörn på fil
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("WritePointsToFile: Something went wrong!!!")
	}
	for i := 0; i < mesh.No_tri; i++ {
		s_x := strconv.FormatFloat(mesh.Triangles[i][0], 'f', 2, 64)
		s_y := strconv.FormatFloat(mesh.Triangles[i][1], 'f', 2, 64)
		s_z := strconv.FormatFloat(mesh.Triangles[i][2], 'f', 2, 64)
		s := s_x + " " + s_y + " " + s_z + "\n"
		file.WriteString(s)

		s_x = strconv.FormatFloat(mesh.Triangles[i][3], 'f', 2, 64)
		s_y = strconv.FormatFloat(mesh.Triangles[i][4], 'f', 2, 64)
		s_z = strconv.FormatFloat(mesh.Triangles[i][5], 'f', 2, 64)
		s = s_x + " " + s_y + " " + s_z + "\n"
		file.WriteString(s)

		s_x = strconv.FormatFloat(mesh.Triangles[i][6], 'f', 2, 64)
		s_y = strconv.FormatFloat(mesh.Triangles[i][7], 'f', 2, 64)
		s_z = strconv.FormatFloat(mesh.Triangles[i][8], 'f', 2, 64)
		s = s_x + " " + s_y + " " + s_z + "\n"
		file.WriteString(s)
	}
}
func (mesh *Mesh) WriteNormalsToFile(path string) {
	// skriver alla triangelnormaler på fil
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("WriteNormalsToFile: Something went wrong!!!")
	}
	for i := 0; i < mesh.No_tri; i++ {
		s_x := strconv.FormatFloat(mesh.Normals[i][0], 'f', 2, 64)
		s_y := strconv.FormatFloat(mesh.Normals[i][1], 'f', 2, 64)
		s_z := strconv.FormatFloat(mesh.Normals[i][2], 'f', 2, 64)
		s := s_x + " " + s_y + " " + s_z + "\n"
		file.WriteString(s)
	}
}
func (mesh *Mesh) WriteProfileToFile(path string) {
	// Skriver Profilen till fil
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("WriteProfileToFile: Something went wrong!!!")
	}
	for i := 0; i < len(mesh.Profile); i++ {
		s_x := strconv.FormatFloat(mesh.Profile[i][0], 'f', 2, 64)
		s_y := strconv.FormatFloat(mesh.Profile[i][1], 'f', 2, 64)
		s := s_x + " " + s_y + "\n"
		file.WriteString(s)
	}
}
func (cs *CrossSection) WriteCrossSectionToFile(path string , cs_index int) {
	write_float := func(value float64) string {
		return strconv.FormatFloat(value, 'f', 2, 64)
	}
	
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		fmt.Println("WriteCrossSectionToFile: Something went wrong!!!")
	}
	for i := range(cs.X[cs_index][:cs.No_cols[cs_index]]) {
		s_x := write_float(cs.X[cs_index][i])
		s_z := write_float(cs.Z[cs_index][i])
		s := s_x + " " + s_z + "\n"
		file.WriteString(s)
	}
}
func (mesh *Mesh) WriteMeshProperties() {
	// skriver alla mesh properties på terminal
	mesh.calculateMeshProperties()
	fmt.Println("Mesh properties:")
	fmt.Printf("width x: %.2f mm\n", mesh.X_max-mesh.X_min)
	fmt.Printf("width y: %.2f mm\n", mesh.Y_max-mesh.Y_min)
	fmt.Printf("width z: %.2f mm\n", mesh.Z_max-mesh.Z_min)
}
func (cs *CrossSection) WriteXYZToFile(path string, mattype string) {
	// skriver X, Y oxh Z matriser i CsMesh på fil
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("WriteXYZToFile: Something went wrong!!!")
	}
	mat := make([][1000]float64, cs.No_rows)
	defer file.Close()
	if mattype == "x" {
		mat = cs.X
	}
	if mattype == "y" {
		mat = cs.Y
	}
	if mattype == "z" {
		mat = cs.Z
	}
	s_row := ""
	for i := 0; i < cs.No_rows; i++ {
		s_row = ""
		for j := 0; j < cs.No_cols[i]; j++ {
			s_row += strconv.FormatFloat(mat[i][j], 'f', 2, 64) + " "
		}
		s_row += "\n"
		file.WriteString(s_row)
	}
}
func (mill *Mill) WriteXYZToFile(path string, mattype string) {
	// skriver X, Y oxh Z matriser i CsMesh på fil
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("WriteXYZToFile: Something went wrong!!!")
	}
	mat := make([][1000]float64, mill.No_rows)
	defer file.Close()
	if mattype == "x" {
		mat = mill.X
	}
	if mattype == "y" {
		mat = mill.Y
	}
	if mattype == "z" {
		mat = mill.Z
	}
	if mattype == "xn" {
		mat = mill.Xn
	}
	if mattype == "yn" {
		mat = mill.Yn
	}
	if mattype == "zn" {
		mat = mill.Zn
	}
	s_row := ""
	for i := 0; i < mill.No_rows; i++ {
		s_row = ""
		for j := 0; j < mill.No_cols[i]; j++ {
			s_row += strconv.FormatFloat(mat[i][j], 'f', 2, 64) + " "
		}
		s_row += "\n"
		file.WriteString(s_row)
	}
}
/*
 * JSON FUNCTIONS
 */
func read_settings(dir string) *Settings {
	// läser in settingsfilen och skriver på Settings strukten
	file, err := os.Open(dir)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	finfo, _ := file.Stat()

	bytes := make([]byte, finfo.Size())
	file.Read(bytes)

	s := new(Settings)
	json.Unmarshal(bytes, &s)
	return s
}
/*
 * FUNCTIONS AND METHODS
 */

/*
 * COMPOSITE FUNCTIONS
 */
func prepare_stl(path string) (*Mesh, *Mesh) {
	board := new(Mesh)
	board.Read(path, 1)
	board.AlignMesh("boardcad")
	board.AlignMeshX()
	board.MoveToCenter2()
	deck, bottom := board.Split()
	bottom.Rotate("y", 180.0)
	return deck, bottom
}
func (mill *Mill) export_to_python(dir string, name string) {
    dir = dir+"/"
    mill.WriteXYZToFile(dir+name+"_zn.txt", "zn")
	mill.WriteXYZToFile(dir+name+"_mx.txt", "x")
    mill.WriteXYZToFile(dir+name+"_my.txt", "y")
    mill.WriteXYZToFile(dir+name+"_mz.txt", "z")
}
/*
 * MAIN FUNCTION
 */
func main() {
	fmt.Println("start...")
	// reads the settings from JSON file
	settings := read_settings("settings.json")
	// prepare the STL files
    stlfile := os.Args[1]
	deck, bottom := prepare_stl(stlfile)
	// calculating the cross sections
    cs_deck := deck.CalculateCrossSections(settings.Yres, 1.0)
    cs_bottom := bottom.CalculateCrossSections(settings.Yres, 1.0)
    tr := settings.ToolRadius
    // --- CALCULATING THE DECK ---
    // calculating the mill coordinates and normals
    mdeck := new(Mill)
    mdeck.CalculateMillCoordinates(cs_deck, settings)
    mdeck.CalculateMillNormals()
    // making mill coordinates as center of the milling tool
    for row:=0; row<mdeck.No_rows; row++ {
		for col:=0; col<mdeck.No_cols[row]; col++ {
			mdeck.X[row][col] += tr*mdeck.Xn[row][col] 
			mdeck.Y[row][col] += tr*mdeck.Yn[row][col] 
			mdeck.Z[row][col] += tr*mdeck.Zn[row][col]
		}
	}
    // --- CALCULATING THE BOTTOM ---
    // calculating the mill coordinates and normals
    mbottom := new(Mill)
    mbottom.CalculateMillCoordinates(cs_bottom, settings)
    mbottom.CalculateMillNormals()
    // making mill coordinates as center of the milling tool
    for row:=0; row<mbottom.No_rows; row++ {
		for col:=0; col<mbottom.No_cols[row]; col++ {
			mbottom.X[row][col] += tr*mbottom.Xn[row][col] 
			mbottom.Y[row][col] += tr*mbottom.Yn[row][col] 
			mbottom.Z[row][col] += tr*mbottom.Zn[row][col]
		}
	}
    
    // export to python
    mdeck.export_to_python("out", "deck")
    mbottom.export_to_python("out", "bottom")
    fmt.Println("done...")
}
