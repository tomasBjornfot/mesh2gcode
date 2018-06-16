package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

// Strukt för hantering av settings från settings.json filen
type Settings struct {
	MachineRotCenter    [2]float64 `json:"MachineRotCenter"`
	MachineLength       float64    `json:"MachineLength"`
	MachineHolderDepth  float64    `json:"MachineHolderDepth"`
	HomingOffset        [3]float64 `json:"HomingOffset"`
	BlockSize           [3]float64 `json:"BlockSize"`
	ToolRadius          float64    `json:"ToolRadius"`
	ToolCuttingLength   float64    `json:"ToolCuttingLength"`
	ToolShaftLength     float64    `json:"ToolShaftLength"`
	XresRough           float64    `json:"XresRough"`
	XresFine            float64    `json:"XresFine"`
	YresRough           float64    `json:"YresRough"`
	YresFine            float64    `json:"YresFine"`
	FeedrateStringer    float64    `json:"FeedrateStringer"`
	FeedrateMax         float64    `json:"FeedrateMax"`
	FeedrateMin         float64    `json:"FeedrateMin"`
	FeedrateChangeLimit float64    `json:"FeedrateChangeLimit"`
	HandlePos           float64    `json:"HandlePos"`
	HandleWidth         int        `json:"HandleWidth"`
	HandleHeightOffset  float64    `json:"HandleHeightOffset"`
	InFolder            string     `json:"InFolder"`
	CamFolder           string     `json:"CamFolder"`
	OutFolder           string     `json:"OutFolder"`
}

// strukt för hantering av STL och Mesher
type Mesh struct {
	triangles [][9]float64
	normals   [][3]float64
	profile   [][2]float64
	no_tri    int
	x_min     float64
	y_min     float64
	z_min     float64
	x_max     float64
	y_max     float64
	z_max     float64
}

// strukt för hantering av cross sections
type CsMesh struct {
	x       [][1000]float64
	y       [][1000]float64
	z       [][1000]float64
	no_rows int
	no_cols [1000]int
}

// strukt för hantering av koordinater till gkoden
type CsMill struct {
	x       [][1000]float64
	y       [][1000]float64
	z       [][1000]float64
	nx      [][1000]float64
	ny      [][1000]float64
	nz      [][1000]float64
	no_rows int
	no_cols int
}

//********* STL OCH MESH HANTERING **********//
func (mesh *Mesh) ReadFromFile(path string, filetype string) {
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
	if filetype == "ascii" {
		mesh.no_tri = (len(s_array)-2)/7 - 1
		mesh.triangles = make([][9]float64, mesh.no_tri)
		mesh.normals = make([][3]float64, mesh.no_tri)
		for i := 1; i < mesh.no_tri+1; i++ {
			s_tri := s_array[7*i+1 : 7*(i+1)+1]
			n_string := strings.Split(s_tri[0], " ")

			// läser in normalerna till triangeln
			for j := 0; j < 3; j++ {
				mesh.normals[row][j], _ = strconv.ParseFloat(n_string[j+2], 32)
			}
			// läser in hörnpunkterna till triangeln
			t1_string := strings.Split(s_tri[2], " ")
			t2_string := strings.Split(s_tri[4], " ")
			t3_string := strings.Split(s_tri[3], " ")
			for j := 0; j < 3; j++ {
				mesh.triangles[row][j+0], _ = strconv.ParseFloat(t1_string[j+3], 32)
				mesh.triangles[row][j+3], _ = strconv.ParseFloat(t2_string[j+3], 32)
				mesh.triangles[row][j+6], _ = strconv.ParseFloat(t3_string[j+3], 32)
			}
			row++
		}
	}
	if filetype == "binary" {
		mesh.no_tri = (len(bytes) - 84) / 50
		mesh.triangles = make([][9]float64, mesh.no_tri)
		mesh.normals = make([][3]float64, mesh.no_tri)
		bits := binary.LittleEndian.Uint32(bytes[0:4])
		var start_index int = 80 // hoppar över header
		start_index += 4         //  hoppar över uint32
		for i := 0; i < mesh.no_tri; i++ {
			// normaler
			for j := 0; j < 3; j++ {
				bits = binary.LittleEndian.Uint32(bytes[(start_index):(start_index + 4)])
				mesh.normals[i][j] = float64(math.Float32frombits(bits))
				start_index += 4
			}
			// trianglar
			for j := 0; j < 9; j++ {
				bits = binary.LittleEndian.Uint32(bytes[(start_index):(start_index + 4)])
				mesh.triangles[i][j] = float64(math.Float32frombits(bits))
				start_index += 4
			}
			// hoppar över uint16
			start_index += 2
		}
	}
}
func (mesh *Mesh) WriteToFile(path string) {
	// skriver en mesh till binär STL fil
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("WriteToFile: Something went wrong!!!")
	}
	defer file.Close()

	header := make([]byte, 80)
	file.Write(header)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, int32(mesh.no_tri))
	for i := 0; i < mesh.no_tri; i++ {
		for j := 0; j < 3; j++ {
			binary.Write(buf, binary.LittleEndian, float32(mesh.normals[i][j]))
		}

		for j := 0; j < 9; j++ {
			binary.Write(buf, binary.LittleEndian, float32(mesh.triangles[i][j]))
		}

		binary.Write(buf, binary.LittleEndian, int16(i))
	}
	buf.WriteTo(file)
}
func (mesh *Mesh) WritePointsToFile(path string) {
	// skriver alla triangelhörn på fil
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("WritePointsToFile: Something went wrong!!!")
	}
	for i := 0; i < mesh.no_tri; i++ {
		s_x := strconv.FormatFloat(mesh.triangles[i][0], 'f', 2, 64)
		s_y := strconv.FormatFloat(mesh.triangles[i][1], 'f', 2, 64)
		s_z := strconv.FormatFloat(mesh.triangles[i][2], 'f', 2, 64)
		s := s_x + " " + s_y + " " + s_z + "\n"
		file.WriteString(s)

		s_x = strconv.FormatFloat(mesh.triangles[i][3], 'f', 2, 64)
		s_y = strconv.FormatFloat(mesh.triangles[i][4], 'f', 2, 64)
		s_z = strconv.FormatFloat(mesh.triangles[i][5], 'f', 2, 64)
		s = s_x + " " + s_y + " " + s_z + "\n"
		file.WriteString(s)

		s_x = strconv.FormatFloat(mesh.triangles[i][6], 'f', 2, 64)
		s_y = strconv.FormatFloat(mesh.triangles[i][7], 'f', 2, 64)
		s_z = strconv.FormatFloat(mesh.triangles[i][8], 'f', 2, 64)
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
	for i := 0; i < mesh.no_tri; i++ {
		s_x := strconv.FormatFloat(mesh.normals[i][0], 'f', 2, 64)
		s_y := strconv.FormatFloat(mesh.normals[i][1], 'f', 2, 64)
		s_z := strconv.FormatFloat(mesh.normals[i][2], 'f', 2, 64)
		s := s_x + " " + s_y + " " + s_z + "\n"
		file.WriteString(s)
	}
}
func (mesh *Mesh) WriteProfileToFile(path string) {
	// Skriver profilen till fil
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("WriteProfileToFile: Something went wrong!!!")
	}
	for i := 0; i < len(mesh.profile); i++ {
		s_x := strconv.FormatFloat(mesh.profile[i][0], 'f', 2, 64)
		s_y := strconv.FormatFloat(mesh.profile[i][1], 'f', 2, 64)
		s := s_x + " " + s_y + "\n"
		file.WriteString(s)
	}
}
func (mesh *Mesh) WriteMeshProperties() {
	// skriver alla mesh properties på terminal
	mesh.CalculateMeshProperties()
	fmt.Println("Mesh properties:")
	fmt.Printf("width x: %.2f mm\n", mesh.x_max-mesh.x_min)
	fmt.Printf("width y: %.2f mm\n", mesh.y_max-mesh.y_min)
	fmt.Printf("width z: %.2f mm\n", mesh.z_max-mesh.z_min)
}
func WriteInfo(mesh *Mesh, csMill *CsMill, setting *Settings) {

	mesh.WriteMeshProperties()
	z_min := 1000.0
	z_max := -1000.0
	for i := 0; i < csMill.no_rows; i++ {
		for j := 0; j < csMill.no_cols; j++ {
			if csMill.z[i][j] < z_min {
				z_min = csMill.z[i][j]
			}
			if csMill.z[i][j] > z_max {
				z_max = csMill.z[i][j]
			}
		}
	}
	fmt.Printf("min depth: %.2f mm\n", z_min)
	fmt.Printf("max depth: %.2f mm\n", z_max)
	fmt.Printf("block size: %.1f mm\n", setting.BlockSize)
}
func (csMesh *CsMesh) WriteXYZToFile(path string, mattype string) {
	// skriver X, Y oxh Z matriser i CsMesh på fil
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("WriteXYZToFile: Something went wrong!!!")
	}
	mat := make([][1000]float64, csMesh.no_rows)
	defer file.Close()
	if mattype == "x" {
		mat = csMesh.x
	}
	if mattype == "y" {
		mat = csMesh.y
	}
	if mattype == "z" {
		mat = csMesh.z
	}
	s_row := ""
	for i := 0; i < csMesh.no_rows; i++ {
		s_row = ""
		for j := 0; j < csMesh.no_cols[i]; j++ {
			s_row += strconv.FormatFloat(mat[i][j], 'f', 2, 64) + " "
		}
		s_row += "\n"
		file.WriteString(s_row)
	}

}
func (csMill *CsMill) WriteXYZToFile(path string, mattype string) {
	// skriver alla X, Y och Z matriser i CsMill på fil
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("WriteXYZToFile: Something went wrong!!!")
	}
	mat := make([][1000]float64, csMill.no_rows)
	defer file.Close()
	if mattype == "x" {
		mat = csMill.x
	}
	if mattype == "y" {
		mat = csMill.y
	}
	if mattype == "z" {
		mat = csMill.z
	}
	if mattype == "nx" {
		mat = csMill.nx
	}
	if mattype == "ny" {
		mat = csMill.ny
	}
	if mattype == "nz" {
		mat = csMill.nz
	}
	s_row := ""
	for i := 0; i < csMill.no_rows; i++ {
		s_row = ""
		for j := 0; j < csMill.no_cols; j++ {
			s_row += strconv.FormatFloat(mat[i][j], 'f', 2, 64) + " "
		}
		s_row += "\n"
		file.WriteString(s_row)
	}

}
func (csMill *CsMill) WriteGcodePointsToFile(setting *Settings, path string, addtoolradius bool) {
	// skriver alla punkter från gkoden till fil
	f, err := os.Create(path)
	if err != nil {
		panic("WriteGcodePointsToFile: Something went wrong!")
	}

	var s string
	cx := csMill.x
	cy := csMill.y
	cz := csMill.z

	if addtoolradius == true {
		for i := 0; i < csMill.no_rows; i++ {
			for j := 0; j < csMill.no_cols; j++ {
				cx[i][j] += setting.ToolRadius * csMill.nx[i][j]
				cy[i][j] += setting.ToolRadius * csMill.ny[i][j]
				cz[i][j] += setting.ToolRadius * csMill.nz[i][j]
			}
		}
	}
	// mittspåret
	for i := 0; i < csMill.no_rows; i++ {
		s += strconv.FormatFloat(cx[i][csMill.no_cols/2], 'f', 2, 64) + ", "
		s += strconv.FormatFloat(cy[i][csMill.no_cols/2], 'f', 2, 64) + ", "
		s += strconv.FormatFloat(cz[i][csMill.no_cols/2], 'f', 2, 64) + "\n"
		f.WriteString(s)
		s = ""
	}
	// en spiral
	center_col := csMill.no_cols / 2
	col := int(0)
	// lägger punkter i en spiral
	for i := 1; i < csMill.no_cols/2+1; i++ {

		col = center_col + i
		for j := 0; j < csMill.no_rows; j++ {
			s += strconv.FormatFloat(cx[j][col], 'f', 2, 64) + ", "
			s += strconv.FormatFloat(cy[j][col], 'f', 2, 64) + ", "
			s += strconv.FormatFloat(cz[j][col], 'f', 2, 64) + "\n"
			f.WriteString(s)
			s = ""
		}

		col = center_col - i
		for j := 0; j < csMill.no_rows; j++ {
			s += strconv.FormatFloat(cx[j][col], 'f', 2, 64) + ", "
			s += strconv.FormatFloat(cy[j][col], 'f', 2, 64) + ", "
			s += strconv.FormatFloat(cz[j][col], 'f', 2, 64) + "\n"
			f.WriteString(s)
			s = ""
		}
	}
}

//********* TRANSLATION OCH ROTATION **********/
func (mesh *Mesh) Rotate(axis string, angle_degrees float64) {
	// roterar en mesh i rymden
	ar := angle_degrees * math.Pi / 180.0
	for i := 0; i < mesh.no_tri; i++ {
		for j := 0; j < 9; j = j + 3 {
			point := mesh.triangles[i][j:(j + 3)]
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
			mesh.triangles[i][j+0] = new_point[0]
			mesh.triangles[i][j+1] = new_point[1]
			mesh.triangles[i][j+2] = new_point[2]
		}
	}
	mesh.CalculateNormals()
}
func (mesh *Mesh) Translate(vector [3]float64) {
	// translaterar en mesh i rymden
	for i := 0; i < mesh.no_tri; i++ {
		for j := 0; j < 9; j = j + 3 {
			mesh.triangles[i][j] = mesh.triangles[i][j] + vector[0]
		}
		for j := 1; j < 9; j = j + 3 {
			mesh.triangles[i][j] = mesh.triangles[i][j] + vector[1]
		}
		for j := 2; j < 9; j = j + 3 {
			mesh.triangles[i][j] = mesh.triangles[i][j] + vector[2]
		}
	}
}
func (csMill *CsMill) TranslateToBlockAndMachine(s *Settings, side string) {
	// gör en translation anpassad till maskinen och blockstorlek
	dx := s.MachineRotCenter[0]
	dy := s.BlockSize[1] / 2.0
	dz := float64(0)
	if side == "deck" {
		dz = s.MachineRotCenter[1] - s.MachineHolderDepth + s.BlockSize[2]/2.0
	}
	if side == "bottom" {
		dz = s.MachineRotCenter[1] + s.MachineHolderDepth - s.BlockSize[2]/2.0
	}
	for i := 0; i < csMill.no_rows; i++ {
		for j := 0; j < csMill.no_cols; j++ {
			csMill.x[i][j] += dx
			csMill.y[i][j] += dy
			csMill.z[i][j] += dz
		}
	}
}

// ********** HELPERS ************
func Linspace(min float64, max float64, no_segments int) []float64 {
	numbers := make([]float64, no_segments+1)
	numbers[0] = min
	segment := (max - min) / float64(no_segments)
	for i := 1; i < len(numbers); i++ {
		numbers[i] = numbers[i-1] + segment
	}
	return numbers
}
func CrossProduct(v0 [3]float64, v1 [3]float64) [3]float64 {
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
func TrianglesToPoints(mesh Mesh) [][3]float64 {
	// tar ut trianglarna från mesh och gör en x,3 matris av punkterna
	points := make([][3]float64, 3*mesh.no_tri)
	// plockar ut punkterna från trianglarna
	for i := 0; i < mesh.no_tri; i++ {
		points[3*i+0][0] = mesh.triangles[i][0]
		points[3*i+1][0] = mesh.triangles[i][3]
		points[3*i+2][0] = mesh.triangles[i][6]

		points[3*i+0][1] = mesh.triangles[i][1]
		points[3*i+1][1] = mesh.triangles[i][4]
		points[3*i+2][1] = mesh.triangles[i][7]

		points[3*i+0][2] = mesh.triangles[i][2]
		points[3*i+1][2] = mesh.triangles[i][5]
		points[3*i+2][2] = mesh.triangles[i][8]
	}
	return points
}
func GetNearestNeighbours(x float64, x_array []float64) []int {
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
func TwoPointsToLine(x0 float64, x1 float64, y0 float64, y1 float64) (float64, float64) {
	k := (y1 - y0) / (x1 - x0)
	m := y0 - k*x0
	return k, m
}
func yValueAt(x float64, k float64, m float64) float64 {
	return k*x + m
}
func GetMinValue(array []float64) int {
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
func GetMaxValue(array []float64) int {
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
func AddSplinePoints(array_x []float64, array_z []float64, no_points int) (x []float64, z []float64) {
	// skapar punkter bestämda efter räta linjer mellan punkterna på array
	array_x_min := array_x[GetMinValue(array_x[:])]
	array_x_max := array_x[GetMaxValue(array_x[:])]
	x = Linspace(array_x_min+0.001, array_x_max-0.001, no_points)
	z = make([]float64, len(x))
	for i := 0; i < len(x); i++ {
		index := GetNearestNeighbours(x[i], array_x)
		k, m := TwoPointsToLine(array_x[index[0]], array_x[index[1]], array_z[index[0]], array_z[index[1]])
		z[i] = yValueAt(x[i], k, m)
		//fmt.Printf("i: %d k: %.2f m: %.2f x: %.2f z: %.2f x0: %.2f x1: %.2f z0: %.2f z1: %.2f index[0] %d index[1] %d\n ", i, k, m, x[i], z[i], array_x[index[0]], array_x[index[1]], array_z[index[0]], array_z[index[1]], index[0], index[1])
	}
	return
}
func (csMesh *CsMesh) GetWidestCs() int {
	// hittar den bredaste cs raden ur csMesh
	index := int(0)
	maxwidth := float64(-1.0)
	for i := 0; i < csMesh.no_rows; i++ {
		imin := GetMinValue(csMesh.x[i][:])
		imax := GetMaxValue(csMesh.x[i][:])
		maxwidth_new := csMesh.x[i][imax] - csMesh.x[i][imin]
		if maxwidth_new > maxwidth {
			maxwidth = maxwidth_new
			index = i
		}
	}
	return index
}
func (csMesh *CsMesh) PatchToPoints(path string) {
	// hittar ingen referens till denna
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("PatchToPoints: Something went wrong!!!")
	}
	defer file.Close()
	s_row := ""
	for i := 0; i < csMesh.no_rows; i++ {
		for j := 0; j < csMesh.no_cols[i]; j++ {
			s_row = ""
			s_row += strconv.FormatFloat(csMesh.x[i][j], 'f', 2, 64) + " "
			s_row += strconv.FormatFloat(csMesh.y[i][j], 'f', 2, 64) + " "
			s_row += strconv.FormatFloat(csMesh.z[i][j], 'f', 2, 64) + "\n"
			file.WriteString(s_row)
		}
	}

}
func (csMill *CsMill) PatchToPoints(path string) {
	// hittar ingen referens till denna
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("PatchToPoints: Something went wrong!!!")
	}
	defer file.Close()
	s_row := ""
	for i := 0; i < csMill.no_rows; i++ {
		for j := 0; j < csMill.no_cols; j++ {
			s_row = ""
			s_row += strconv.FormatFloat(csMill.x[i][j], 'f', 2, 64) + " "
			s_row += strconv.FormatFloat(csMill.y[i][j], 'f', 2, 64) + " "
			s_row += strconv.FormatFloat(csMill.z[i][j], 'f', 2, 64) + "\n"
			file.WriteString(s_row)
		}
	}

}
func (csMesh *CsMesh) GetIndexForSpline(maxdistance float64) []int {
	// vad gör denna???
	// hämtar det bredaste stället på brädan
	index_row := csMesh.GetWidestCs()
	x_spline := csMesh.x[index_row][0:csMesh.no_cols[index_row]]
	z_spline := csMesh.z[index_row][0:csMesh.no_cols[index_row]]
	// lägger till en massa punkter
	x, z := AddSplinePoints(x_spline, z_spline, 5001)

	// hittar index till punkter som ligger närmast max avståndet (maxdistance)
	// det är detts punkter som är fräsbanorna för tool radius 0

	// beräknar alla avstånd mellan x,z'na
	d := make([]float64, len(x)-1)
	for i := 0; i < len(d); i++ {
		dx := x[i+1] - x[i]
		dz := z[i+1] - z[i]
		d[i] = math.Sqrt(dx*dx + dz*dz)
	}
	// tar ut index inom avstånd maxdistance
	// börjar från mitten och går i positiv rikning
	dist := float64(0)
	index_r := make([]int, 1000)
	counter := int(0)
	for i := len(d) / 2; i < len(d); i++ {
		dist += d[i]
		if dist > maxdistance {
			//fmt.Println("counter", counter, "i:", i, "dist: ",dist)
			dist = 0
			//i--
			index_r[counter] = i
			counter++
		}
	}
	// tar med det sista värdet (om det inte redan är med)
	if index_r[counter-1] != 5000 {
		index_r[counter] = 5000
		counter++
	}

	// spegelvänder alla index runt centrum punkten
	index_final := make([]int, 2*counter+1)
	for i := 0; i < counter; i++ {
		index_final[i] = len(d) - 1 - index_r[counter-1-i]
	}
	index_final[counter] = len(d) / 2
	for i := counter + 1; i < 2*counter+1; i++ {
		index_final[i] = index_r[i-counter-1]
	}
	return index_final
}

// special funktioner
func (mesh *Mesh) CalculateNormals() {
	// räknar ut normaler för trianglar
	var v0, v1 [3]float64
	for i := 0; i < mesh.no_tri; i++ {
		v0[0] = mesh.triangles[i][0] - mesh.triangles[i][6]
		v0[1] = mesh.triangles[i][1] - mesh.triangles[i][7]
		v0[2] = mesh.triangles[i][2] - mesh.triangles[i][8]

		v1[0] = mesh.triangles[i][0] - mesh.triangles[i][3]
		v1[1] = mesh.triangles[i][1] - mesh.triangles[i][4]
		v1[2] = mesh.triangles[i][2] - mesh.triangles[i][5]

		mesh.normals[i] = CrossProduct(v0, v1)
	}
}
func (mesh *Mesh) MoveToCenter() {
	// flyttar en mesh till centrum
	var x_sum, y_sum, z_sum float64 = 0, 0, 0
	var vector [3]float64
	for i := 0; i < mesh.no_tri; i++ {
		x_sum += mesh.triangles[i][0]
		x_sum += mesh.triangles[i][3]
		x_sum += mesh.triangles[i][6]

		y_sum += mesh.triangles[i][1]
		y_sum += mesh.triangles[i][4]
		y_sum += mesh.triangles[i][7]

		z_sum += mesh.triangles[i][2]
		z_sum += mesh.triangles[i][5]
		z_sum += mesh.triangles[i][8]
	}
	vector[0] = -x_sum / (3 * float64(mesh.no_tri))
	vector[1] = -y_sum / (3 * float64(mesh.no_tri))
	vector[2] = -z_sum / (3 * float64(mesh.no_tri))
	fmt.Println("translation vector: ", vector)
	mesh.Translate(vector)
}
func (mesh *Mesh) MoveToCenter2() {
	// flyttar till centrum map högsta och minsta värde
	var x_min, y_min, z_min float64 = 1000.0, 1000.0, 1000.0
	var x_max, y_max, z_max float64 = -1000.0, -1000.0, -1000.0
	var vector [3]float64
	for i := 0; i < mesh.no_tri; i++ {
		for j := 0; j < 7; j = j + 3 {
			if mesh.triangles[i][j] < x_min {
				x_min = mesh.triangles[i][j]
			}
			if mesh.triangles[i][j] > x_max {
				x_max = mesh.triangles[i][j]
			}
		}
		for j := 1; j < 8; j = j + 3 {
			if mesh.triangles[i][j] < y_min {
				y_min = mesh.triangles[i][j]
			}
			if mesh.triangles[i][j] > y_max {
				y_max = mesh.triangles[i][j]
			}
		}
		for j := 2; j < 9; j = j + 3 {
			if mesh.triangles[i][j] < z_min {
				z_min = mesh.triangles[i][j]
			}
			if mesh.triangles[i][j] > z_max {
				z_max = mesh.triangles[i][j]
			}
		}
	}
	vector[0] = -(x_max + x_min) / 2.0
	vector[1] = -(y_max + y_min) / 2.0
	vector[2] = -(z_max + z_min) / 2.0
	fmt.Println("translation vector: ", vector)
	mesh.Translate(vector)
}
func (mesh *Mesh) AlignMesh(cadtype string) {
	// gör en alignment beroende på vilkan cad som används
	if cadtype == "boardcad" {
		mesh.MoveToCenter()
		mesh.Rotate("x", 90)
		mesh.Rotate("z", 90)
	}
	mesh.CalculateNormals()
}
func (mesh *Mesh) AlignMeshX() {
	// roterar brädan runt x vectorn tills man hittar ett minimum z_max - z_min
	mesh.CalculateMeshProperties()
	rmesh := mesh
	z_range := rmesh.z_max - rmesh.z_min
	for i := 0; i < 50; i++ {
		rmesh.Rotate("x", -0.1)
		rmesh.CalculateMeshProperties()
		if rmesh.z_max-rmesh.z_min < z_range {
			z_range = rmesh.z_max - rmesh.z_min
		} else {
			//fmt.Printf("Alignment x rotation: %0.2f degrees\n", 0.1*float64(i))
			break
		}
	}
	mesh = rmesh
}
func (mesh *Mesh) CalculateMeshProperties() {
	// tar reda på max och min i varje dimension
	mesh.x_min, mesh.y_min, mesh.z_min = 100000.0, 100000.0, 100000.0
	mesh.x_max, mesh.y_max, mesh.z_max = -100000.0, -100000.0, -100000.0
	points := TrianglesToPoints(*mesh)

	for i := 0; i < 3*mesh.no_tri; i++ {
		// min x
		if points[i][0] < mesh.x_min {
			mesh.x_min = points[i][0]
		}
		// min y
		if points[i][1] < mesh.y_min {
			mesh.y_min = points[i][1]
		}
		// min z
		if points[i][2] < mesh.z_min {
			mesh.z_min = points[i][2]
		}
		// max x
		if points[i][0] > mesh.x_max {
			mesh.x_max = points[i][0]
		}
		// max y
		if points[i][1] > mesh.y_max {
			mesh.y_max = points[i][1]
		}
		// max z
		if points[i][2] > mesh.z_max {
			mesh.z_max = points[i][2]
		}
	}
}
func (mesh *Mesh) Split() (*Mesh, *Mesh) {
	// delar upp deck och bottom på brädan
	// flytta funktionen
	no_tri_deck := int(0)
	no_tri_bottom := int(0)
	for i := 0; i < mesh.no_tri; i++ {
		if mesh.normals[i][2] < 0 {
			no_tri_deck++
		}
		if mesh.normals[i][2] >= 0 {
			no_tri_bottom++
		}
	}
	deck := new(Mesh)
	deck.triangles = make([][9]float64, no_tri_deck)
	deck.normals = make([][3]float64, no_tri_deck)
	deck.no_tri = no_tri_deck

	bottom := new(Mesh)
	bottom.triangles = make([][9]float64, no_tri_bottom)
	bottom.normals = make([][3]float64, no_tri_bottom)
	bottom.no_tri = no_tri_bottom

	i_deck := int(0)
	i_bottom := int(0)
	for i := 0; i < mesh.no_tri; i++ {
		if mesh.normals[i][2] < 0 {
			deck.triangles[i_deck] = mesh.triangles[i]
			deck.normals[i_deck] = mesh.normals[i]
			i_deck++
		}
		if mesh.normals[i][2] >= 0 {
			bottom.triangles[i_bottom] = mesh.triangles[i]
			bottom.normals[i_bottom] = mesh.normals[i]
			i_bottom++
		}
	}
	return deck, bottom
}
func (mesh *Mesh) CalculateProfile(radius float64, resolution int) {
	// räknar ut profilen på brädan i xy planet

	// plockar ut punkterna från trianglarna
	points := TrianglesToPoints(*mesh)

	x := make([]float64, len(points))
	y := make([]float64, len(points))
	for i := 0; i < len(points); i++ {
		x[i] = points[i][1]
		y[i] = points[i][0]
	}

	// tar reda på y värden på profilen genom att:
	// 0. skapa x värden som cikeln ska ramla ner på
	// 1. iterera över alla x drop värden
	// 2. iterera över alla punkter
	// 3. väljer bara punkter där y > 0
	// 4. tar ut dom punkter som ligger inom [x - radius,x + radius]
	// 5. beräkna höjd på cirkel som krockar med värdet
	// 6. ta ut index på punkten (som cirkeln krockar med)
	// 7. om det finns ett stötte värde, gäller detta

	// 0
	mesh.CalculateMeshProperties()
	drop_x := Linspace(mesh.y_min+0.1-radius, mesh.y_max-0.1+radius, resolution)
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
	mesh.profile = make([][2]float64, no_index)
	pindex := int(0)
	for i := 0; i < len(index); i++ {
		if index[i] != -1 {
			mesh.profile[pindex][0] = x[index[i]]
			mesh.profile[pindex][1] = y[index[i]]
			pindex++
		}
	}
}
func (mesh *Mesh) CalculateCS_Y_Values(max_distance float64, resolution float64) []float64 {
	// vad gör denna. Dåligt namn på funktionen?
	// hämtar profilen
	px := make([]float64, len(mesh.profile))
	py := make([]float64, len(mesh.profile))
	for i := 0; i < len(mesh.profile); i++ {
		px[i] = mesh.profile[i][0]
		py[i] = mesh.profile[i][1]
	}
	// skapar "cross sections" cs_x och cs_y
	cs_x := make([]float64, 100000)
	cs_y := make([]float64, 100000)
	start_index := GetMinValue(px)
	stop_index := GetMaxValue(px)
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
		nindex = GetNearestNeighbours(cs_x_new, px)
		// räknar ut y värdet
		k, m = TwoPointsToLine(px[nindex[0]], px[nindex[1]], py[nindex[0]], py[nindex[1]])
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

//********* CROSS SECTION **********/
func (csMesh *CsMesh) MeshToCs(cs []float64, mesh *Mesh) {
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
		for j := 0; j < mesh.no_tri; j++ {
			if mesh.triangles[j][1]-cs[i] > 0 {
				side[0] = 1
			} else {
				side[0] = -1
			}
			if mesh.triangles[j][4]-cs[i] > 0 {
				side[1] = 1
			} else {
				side[1] = -1
			}
			if mesh.triangles[j][7]-cs[i] > 0 {
				side[2] = 1
			} else {
				side[2] = -1
			}
			// om y korsar triangeln
			side_sum = side[0] + side[1] + side[2]
			if side_sum == 1 || side_sum == -1 {
				tri = mesh.triangles[j]
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
	}
	csMesh.no_cols = no_cols
	csMesh.no_rows = len(cs)
	csMesh.x = x
	csMesh.z = z

	for i := 0; i < len(cs); i++ {
		for j := 0; j < 1000; j++ {
			y[i][j] = cs[i]
		}
	}
	csMesh.y = y
}

//********* Mill COORDINATES **********//
func (csMill *CsMill) CsSplineToMill(csMesh *CsMesh, maxdistance float64) {
	// bekrivning!!
	// byt namn till CsMeshToCsMill??

	// tar ut vilka index som ska sparas av splinen för alla rader
	spline_index := csMesh.GetIndexForSpline(maxdistance)
	x := make([][1000]float64, csMesh.no_rows)
	y := make([][1000]float64, csMesh.no_rows)
	z := make([][1000]float64, csMesh.no_rows)
	for i := 0; i < csMesh.no_rows; i++ {
		//for i:=csMesh.no_rows - 1; i<csMesh.no_rows; i++ {
		x_spline_in := csMesh.x[i][0:csMesh.no_cols[i]]
		z_spline_in := csMesh.z[i][0:csMesh.no_cols[i]]
		x_spline, z_spline := AddSplinePoints(x_spline_in, z_spline_in, 5001)
		for j := 0; j < len(spline_index); j++ {
			x[i][j] = x_spline[spline_index[j]]
			y[i][j] = csMesh.y[i][0]
			z[i][j] = z_spline[spline_index[j]]
		}
	}
	csMill.no_rows = csMesh.no_rows
	csMill.no_cols = len(spline_index)
	csMill.x = x
	csMill.y = y
	csMill.z = z
}
func (csMill *CsMill) CalculateMillNormals() {
	// bekrivning!
	// byt namn till CalculateCsMillNormals
	nx := make([][1000]float64, csMill.no_rows)
	ny := make([][1000]float64, csMill.no_rows)
	nz := make([][1000]float64, csMill.no_rows)
	var v0 [3]float64
	var v1 [3]float64
	var n [3]float64
	// beräknar alla utom kanterna
	for i := 1; i < csMill.no_rows-1; i++ {
		for j := 1; j < csMill.no_cols-1; j++ {
			v0[0] = csMill.x[i][j+1] - csMill.x[i][j-1]
			v0[1] = csMill.y[i][j+1] - csMill.y[i][j-1]
			v0[2] = csMill.z[i][j+1] - csMill.z[i][j-1]

			v1[0] = csMill.x[i+1][j] - csMill.x[i-1][j]
			v1[1] = csMill.y[i+1][j] - csMill.y[i-1][j]
			v1[2] = csMill.z[i+1][j] - csMill.z[i-1][j]
			n = CrossProduct(v0, v1)
			l := math.Sqrt(n[0]*n[0] + n[1]*n[1] + n[2]*n[2])
			nx[i][j] = n[0] / l
			ny[i][j] = n[1] / l
			nz[i][j] = n[2] / l
		}
	}
	// första och sista kolumnen
	for i := 0; i < csMill.no_rows; i++ {
		// tar grannens vaäde
		nx[i][0] = nx[i][1]
		ny[i][0] = ny[i][1]
		nz[i][0] = nz[i][1]

		nx[i][csMill.no_cols-1] = nx[i][csMill.no_cols-2]
		ny[i][csMill.no_cols-1] = ny[i][csMill.no_cols-2]
		nz[i][csMill.no_cols-1] = nz[i][csMill.no_cols-2]
		/*
			nx[i][0] = -1.0
			ny[i][0] = 0.0
			nz[i][0] = 0.0

			nx[i][csMill.no_cols - 1] = 1.0
			ny[i][csMill.no_cols - 1] = 0.0
			nz[i][csMill.no_cols - 1] = 0.0
		*/
	}
	// första och sista raden
	for i := 0; i < csMill.no_cols; i++ {
		// tar grannens värde
		nx[0][i] = nx[1][i]
		ny[0][i] = ny[1][i]
		nz[0][i] = nz[1][i]

		nx[csMill.no_rows-1][i] = nx[csMill.no_rows-2][i]
		ny[csMill.no_rows-1][i] = ny[csMill.no_rows-2][i]
		nz[csMill.no_rows-1][i] = nz[csMill.no_rows-2][i]
		/*
			nx[0][i] = 0.0
			ny[0][i] = -1.0
			nz[0][i] = 0.0

			nx[csMill.no_rows - 1][i] = 0.0
			ny[csMill.no_rows - 1][i] = 1.0
			nz[csMill.no_rows - 1][i] = 0.0
		*/
	}

	csMill.nx = nx
	csMill.ny = ny
	csMill.nz = nz
}

func (csMill *CsMill) MillToEval(path string) {
	s := ""

	// opens a file
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("MillToEval: Something went wrong!!!")
	}
	defer file.Close()

	for i := 0; i < csMill.no_cols; i++ {
		for j := 0; j < csMill.no_rows; j++ {
			s = ""
			s += strconv.FormatFloat(csMill.x[j][i], 'f', 2, 64) + " "
			s += strconv.FormatFloat(csMill.y[j][i], 'f', 2, 64) + " "
			s += strconv.FormatFloat(csMill.z[j][i], 'f', 2, 64) + "\n"
			file.WriteString(s)
		}
	}
}

func (csMill *CsMill) MillToGcode(path string, setting *Settings) {

	// lägger på normal och millradius
	cx := csMill.x
	cy := csMill.y
	cz := csMill.z

	for i := 0; i < csMill.no_rows; i++ {
		for j := 0; j < csMill.no_cols; j++ {
			cx[i][j] = csMill.x[i][j] + setting.ToolRadius*csMill.nx[i][j]
			cy[i][j] = csMill.y[i][j] + setting.ToolRadius*csMill.ny[i][j]
			cz[i][j] = csMill.z[i][j] + setting.ToolRadius*csMill.nz[i][j]
		}
	}

	var s string

	// öppnar en fil
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("MillToGcode: Something went wrong!!!")
	}
	defer file.Close()

	// lägger på en punkt ovanför första kooordinaten
	s += "G1"
	s += " X" + strconv.FormatFloat(cx[0][csMill.no_cols/2], 'f', 2, 64)
	s += " Y" + strconv.FormatFloat(cy[0][csMill.no_cols/2], 'f', 2, 64)
	s += " Z0"
	s += " F" + strconv.FormatFloat(setting.FeedrateStringer, 'f', 2, 64) + "\n"
	file.WriteString(s)
	s = ""

	// lägger punker för fräsning av stringer
	for i := 0; i < csMill.no_rows; i++ {
		s += "G1"
		s += " X" + strconv.FormatFloat(cx[i][csMill.no_cols/2], 'f', 2, 64)
		s += " Y" + strconv.FormatFloat(cy[i][csMill.no_cols/2], 'f', 2, 64)
		s += " Z" + strconv.FormatFloat(cz[i][csMill.no_cols/2], 'f', 2, 64)
		s += " F" + strconv.FormatFloat(setting.FeedrateStringer, 'f', 0, 64) + "\n"
		file.WriteString(s)
		s = ""
	}
	center_col := csMill.no_cols / 2
	col := int(0)

	// räknar ut vilken Feedrate varje rad ska ha
	// tror att det är lätt om:
	//	man gör en vektor med längden csMill.no_rows som heter fr
	fr := make([]float64, csMill.no_rows)
	fcl := setting.FeedrateChangeLimit
	f_max := setting.FeedrateMax
	f_min := setting.FeedrateMin
	fc := f_max - f_min
	y := make([]float64, csMill.no_rows)
	for i := 0; i < csMill.no_rows; i++ {
		y[i] = csMill.y[i][0]
	}
	y_min := y[GetMinValue(y[:])]
	y_max := y[GetMaxValue(y[:])]

	// Gör en vektor som ger en Feedrate för varje y värde
	//  Nära toppen: f_min + fc*(y_max - yn)/d
	//  Nära botten: f_min + fc*(yn - y_min)/d
	//  Mitt på brädan: f_max
	for j := 0; j < csMill.no_rows; j++ {
		if y[j] > y_max-fcl {
			fr[j] = f_min + fc*(y_max-y[j])/fcl
		}
		if csMill.y[j][0] < y_min+fcl {
			fr[j] = f_min + fc*(y[j]-y_min)/fcl
		}
		if y[j] >= y_min+fcl && y[j] <= y_max-fcl {
			fr[j] = f_max
		}
	}

	// lägger punkter i en spiral
	for i := 1; i < csMill.no_cols/2+1; i++ {
		col = center_col + i
		for j := csMill.no_rows - 1; j > -1; j-- {
			s += "G1"
			s += " X" + strconv.FormatFloat(cx[j][col], 'f', 2, 64)
			s += " Y" + strconv.FormatFloat(cy[j][col], 'f', 2, 64)
			s += " Z" + strconv.FormatFloat(cz[j][col], 'f', 2, 64)
			s += " F" + strconv.FormatFloat(fr[j], 'f', 0, 64) + "\n"
			file.WriteString(s)
			s = ""
		}

		col = center_col - i
		for j := 0; j < csMill.no_rows; j++ {
			s += "G1"
			s += " X" + strconv.FormatFloat(cx[j][col], 'f', 2, 64)
			s += " Y" + strconv.FormatFloat(cy[j][col], 'f', 2, 64)
			s += " Z" + strconv.FormatFloat(cz[j][col], 'f', 2, 64)
			s += " F" + strconv.FormatFloat(fr[j], 'f', 0, 64) + "\n"
			file.WriteString(s)
			s = ""
		}

	}
}

// ********** HANDTAG ***********//
func (csMill *CsMill) Add_handles(position float64, width int, h_offset float64) {
	// beskrivning
	// byt namn till AddHandles
	start := int(float64(csMill.no_rows)*position) - width
	end := int(float64(csMill.no_rows)*position) + width
	max_h := csMill.z[start][GetMaxValue(csMill.z[start+width][0:csMill.no_cols])]
	for i := start; i < (end + 1); i++ {
		for j := 0; j < csMill.no_cols; j++ {
			if csMill.z[i][j] < (max_h - h_offset) {
				csMill.z[i][j] = max_h - h_offset
			}
		}

	}

	start = int(float64(csMill.no_rows)*(1-position)) - width
	end = int(float64(csMill.no_rows)*(1-position)) + width
	max_h = csMill.z[start][GetMaxValue(csMill.z[start+width][0:csMill.no_cols])]
	for i := start; i < (end + 1); i++ {
		for j := 0; j < csMill.no_cols; j++ {
			if csMill.z[i][j] < (max_h - h_offset) {
				csMill.z[i][j] = max_h - h_offset
			}
		}

	}
}
func (csMill *CsMill) CalculateHandleNormals(position float64, width int) {
	// Sätter normalerna på pinnarna alltid rakt upp
	// byt namn CalculateHandleNormals

	// pinnarna uppe
	start := int(float64(csMill.no_rows)*position) - width
	end := int(float64(csMill.no_rows)*position) + width
	for i := start; i < (end + 1); i++ {
		csMill.nx[i][0] = 0.0
		csMill.ny[i][0] = 0.0
		csMill.nz[i][0] = 1.0
		csMill.nx[i][csMill.no_cols-1] = 0.0
		csMill.ny[i][csMill.no_cols-1] = 0.0
		csMill.nz[i][csMill.no_cols-1] = 1.0
	}
	//pinnarna nere
	start = int(float64(csMill.no_rows)*(1-position)) - width
	end = int(float64(csMill.no_rows)*(1-position)) + width
	for i := start; i < (end + 1); i++ {
		for j := 0; j < csMill.no_cols; j++ {
			csMill.nx[i][0] = 0.0
			csMill.ny[i][0] = 0.0
			csMill.nz[i][0] = 1.0
			csMill.nx[i][csMill.no_cols-1] = 0.0
			csMill.ny[i][csMill.no_cols-1] = 0.0
			csMill.nz[i][csMill.no_cols-1] = 1.0
		}
	}
}

//********** JSON ***********//
func GetJsonSettings(dir string) *Settings {
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

func WritePropertiesToJsonFile(path string) {
	// skriver properties på en json fil
	type Size struct {
		Xsize float64
		Ysize float64
		Zsize float64
	}
	size := &Size{Xsize: 23.6 - 11.44, Ysize: 44.6, Zsize: 55.1}
	bSize, _ := json.MarshalIndent(size, "", "\t")

	file, err := os.Create(path)
	if err != nil {
		fmt.Println("WritePropertiesToJsonFile: Something went wrong!")
	}
	file.WriteString(string(bSize))
	file.Close()
}

//********** SAMMANSATTA FUNKTIONER ***********//
func (csMill *CsMill) CalculateSide(cut string, side string, mesh *Mesh, setting *Settings) {
	// en sammansatt funktion som gör alla beräkningar för en sida
	mesh.CalculateProfile(50.0, 100)

	cs_mesh := mesh.CalculateCS_Y_Values(setting.YresFine, 1)
	if cut == "rough" {
		cs_mesh = mesh.CalculateCS_Y_Values(setting.YresRough, 1)
	}
	csMesh := new(CsMesh)
	csMesh.MeshToCs(cs_mesh, mesh)
	csMesh.WriteXYZToFile(setting.OutFolder+"_"+side+"_x", "x")
	csMesh.WriteXYZToFile(setting.OutFolder+"_"+side+"_y", "y")
	csMesh.WriteXYZToFile(setting.OutFolder+"_"+side+"_z", "z")
	if cut == "fine" {
		csMill.CsSplineToMill(csMesh, setting.XresFine)
	}
	if cut == "rough" {
		csMill.CsSplineToMill(csMesh, setting.XresRough)
	}
	if side == "deck" {
		csMill.Add_handles(setting.HandlePos, setting.HandleWidth, setting.HandleHeightOffset)
	}
	csMill.CalculateMillNormals()
	if side == "deck" {
		csMill.CalculateHandleNormals(setting.HandlePos, setting.HandleWidth)
	}
}

// ************** FILHANTERING **********************
func GetFilesFromDir(dir string) ([]string, int) {

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	filestring := make([]string, 20)
	no_files := int(0)
	for _, f := range files {
		fn := f.Name()
		if fn[(len(fn)-4):] == ".stl" || fn[(len(fn)-4):] == ".STL" {
			filestring[no_files] = f.Name()
			no_files++
		}
	}
	return filestring, no_files
}

//********** MAIN ***********//
func main() {

	// gör en funktion för att kunna använda go-routine på bästa sätt
	Calculate := func(stlfile string, cut string, side string, mesh *Mesh, setting *Settings) {

		// räknar ut csMill
		csMill := new(CsMill)
		csMill.CalculateSide(cut, side, mesh, setting)
		//mesh.WriteProfileToFile(setting.OutFolder + side + "_profile")
		csMill.TranslateToBlockAndMachine(setting, side)

		// skriver lite extra filer
		//csMill.WriteXYZToFile(setting.OutFolder+stlfile[0:len(stlfile)-4]+"_"+side+"_csMill_x", "x")
		//csMill.WriteXYZToFile(setting.OutFolder+stlfile[0:len(stlfile)-4]+"_"+side+"_csMill_y", "y")
		//csMill.WriteXYZToFile(setting.OutFolder+stlfile[0:len(stlfile)-4]+"_"+side+"_csMill_z", "z")

		// skriver ut eval filen
		savefileEval := setting.OutFolder + stlfile[0:len(stlfile)-4] + "_" + side + "_eval"
		csMill.MillToEval(savefileEval)
		// skriver ut gkod filen
		//savefileMill := setting.CamFolder + stlfile[0:len(stlfile)-4] + "_" + side + "_" + cut + ".gc"
		//savefileMill := setting.CamFolder + stlfile[0:len(stlfile)-4] + "_" + side + ".gc"
		savefileMill := setting.CamFolder + side + ".gc"
		csMill.MillToGcode(savefileMill, setting)
		fmt.Println(side + " " + cut)
		WriteInfo(mesh, csMill, setting)
		fmt.Println()
	}

	// tar tid
	start := time.Now()

	// läser in settings filen
	setting := GetJsonSettings("./settings.json")
	fmt.Println(setting.InFolder)
	// hämtar filnamen på alla filer i share folder
	// files, no_files := GetFilesFromDir(settInFolder)

	stlfile := os.Args[1]
	fmt.Println("mesh2gcode:", stlfile)

	// STL till mesh
	mesh := new(Mesh)
	mesh.ReadFromFile(setting.InFolder+stlfile, "ascii")

	// Alignment beroende på vilken CAD som används
	mesh.AlignMesh("boardcad")

	// Roterar för att minimera materialtjocklek
	mesh.AlignMeshX()
	mesh.MoveToCenter2()

	// delar på botten och toppen
	deck, bottom := mesh.Split()
	bottom.Rotate("y", 180.0)

	// räknar ut gkod fil för bottom
	// Calculate(stlfile, "rough", "bottom", bottom, setting)
	Calculate(stlfile, "fine", "bottom", bottom, setting)

	// räknar ut gkod fil för deck
	// Calculate(stlfile, "rough", "deck", deck, setting)
	Calculate(stlfile, "fine", "deck", deck, setting)

	//
	fmt.Println()

	// skriver ut lite extra filer
	//mesh.WriteToFile(setting.OutFolder + stlfile[:(len(stlfile)-4)] + "_mesh.stl")
	//deck.WriteToFile(setting.OutFolder + stlfile[:(len(stlfile)-4)] + "_deck.stl")
	//bottom.WriteToFile(setting.OutFolder + stlfile[:(len(stlfile)-4)] + "_bottom.stl")
	// tar tiden
	t := time.Now()
	fmt.Println("Done in:", t.Sub(start))
}
