package main

import (
    "fmt"
    "os"
    "strconv"
    "encoding/json"
)

/*
 * STRUCTS
 */
type Settings struct {
    ToolRadius float64 `json:"ToolRadius"`
    MaxHeight float64 `json:"MaxHeight"`
    Xpos    float64 `json:"Xpos"`
    Ypos    float64 `json:"Ypos"`
    MaterialThickness  float64 `json:"MaterialThickness"`
    FilePrefix  string `json:"FilePrefix"`
}
/*
 * FUNCTIONS
 */
func read_json_settings() *Settings {
	// läser in settingsfilen och skriver på Settings strukten
	file, err := os.Open("calibration.json")
	if err != nil {
		fmt.Println("Something went wrong when reading calibration.json file...")
	}
	defer file.Close()
	finfo, _ := file.Stat()

	bytes := make([]byte, finfo.Size())
	file.Read(bytes)

	s := new(Settings)
	json.Unmarshal(bytes, &s)
	return s
}
func make_plane(start [3]float64, end [3]float64, dx float64, no_lines int) [][3]float64 {
    lines := make([][3]float64, 2*no_lines)
    for i := 0; i < no_lines; i++ {
        if i%2 == 0 {
            lines[2*i] = [3]float64{start[0] + float64(i)*dx, start[1], start[2]}
            lines[2*i+1] = [3]float64{end[0] + float64(i)*dx, end[1], end[2]}
        }
        if i%2 == 1 {
            lines[2*i+1] = [3]float64{start[0] + float64(i)*dx, start[1], start[2]}
            lines[2*i] = [3]float64{end[0] + float64(i)*dx, end[1], end[2]}
        }
    }
    return lines
}

func make_planes(plane [][3]float64, dz float64, no_planes int) [][3]float64{
    
    new_plane := make([][3]float64, len(plane))
    copy(new_plane, plane)
    planes := make([][3]float64, len(plane))
    copy(planes, plane)
    
    for i:=1; i<no_planes ; i++ {
        for j:= range new_plane {
            new_plane[j][2] -= dz 
        }
        planes = append(planes, new_plane...)
    }
    return planes
}

func make_gcode(filename string, data [][3]float64) {
    s := ""
    for i := 0; i < len(data); i++ {
        x_string := strconv.FormatFloat(data[i][0], 'f', 1, 64)
        y_string := strconv.FormatFloat(data[i][1], 'f', 1, 64)
        z_string := strconv.FormatFloat(data[i][2], 'f', 1, 64)
        s += "G1"
        s += " X" + x_string
        s += " Y" + y_string
        s += " Z" + z_string
        s += " F1300"
        s += "\n"
    }
    myfile, _ := os.Create(filename)
    myfile.WriteString(s)
}
func add_start_point(plane [][3]float64, max_height float64) [][3]float64 {
    start_point := make([][3]float64, 1)
    start_point[0] = [3]float64 {plane[0][0], plane[0][1], max_height}
    return append(start_point, plane...)
}
func add_end_point(plane [][3]float64, max_height float64) [][3]float64 {
    end_point := make([][3]float64, 1)
    end_point[0] = [3]float64 {plane[len(plane)-1][0], plane[len(plane)-1][1], max_height}
    return append(plane, end_point...)
}
/*
 * THE MAIN FUNCTION
 */
func main() {
    // indata
    s := read_json_settings()
    tr := s.ToolRadius
    max_height := s.MaxHeight // same as HomeOffset[2]
    xpos := s.Xpos // avvikelse från centrum
    ypos := s.Ypos // nedre y position
    file_prefix := s.FilePrefix //
    
    // make the wide plane for the horizontal (z) plane
    p0 := [3]float64 {xpos, ypos, 40 + tr}
    p1 := [3]float64 {xpos, ypos + 50, 40 + tr}
    plane_z := make_plane(p0, p1, 5.0, 16)
    
    // adds a start point to plane z
    plane_z = add_start_point(plane_z, max_height)

    // make the smaller planes to make the vertical (x) plane
    p0 = [3]float64 {xpos, ypos, 30 + tr}
    p1 = [3]float64 {xpos, ypos + 50, 30 + tr}
    plane_x := make_plane(p0, p1, 5.0, 8) 
    planes_x := make_planes(plane_x, 10.0, 5) 
    
    // adds an end point to plane x 
    planes_x = add_end_point(planes_x, max_height)

    // merge plane_z and plane_x
    planes0 := append(plane_z, planes_x...)

    // make the other side 
    planes1 := make([][3]float64, len(planes0))
    copy(planes1, planes0)
    for i := range planes1 {
        planes1[i][0] = -planes1[i][0]
    } 
    planes := append(planes0, planes1...)
    
    // adds an end point to planes
    planes = add_end_point(planes, max_height)
    
    make_gcode(file_prefix+"_calibration.gc", planes)
    
}
