# mesh2gcode
A program to calculate g code data from an STL file.

# The following functions are implemented
## StlData.readAsciiStl(string)
Reads an ascii stl file and returns normals and verticies.
## StlData.writeBinaryStl(string)
Writes a binary stl file from StlData struct.
## StlData.alignStl(string)
Aligns the stl data depending on CAD. Only BoardCAD is implemented.
The alignment is done in two steps:
* Move to 0,0,0, rotate so y is aling the board and z+ is on the deck
* Rotate around x axis to minimize the board height
## StlData.calculateProperties()
Calculate the min, max values for x, y and z plus length, width and height.
## StlData.splitStl()
Split the stl file to a deck and bottom 
## Gcode.calculateProfile()
Calculates the profile on the board in the xy plane
## Gcode.calculateYdata(float64)
Calculates the y coordinates for the g-code file.
The y coordinate can be seen as a cross-section of the board along the y axis 
The resolution is depends on the input as the maximum allowed distance between y coordinates on the profile 
Returns a matrix (patch) of y values
## Gcode.calculateXZdata(float64)
Calculates the x and z coordinates for the g-code file by using cross-sections of the stl file
Returns matrises (patches) of y and z  values
## Gcode.calculatePath()
Calculates the path from the patches above
