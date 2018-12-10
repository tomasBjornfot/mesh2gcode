# A package for handling of stl files.

## Public functions:
### points, normals, err =  Read(filename string, filetype int)
filename => the full path of the stl file
filetype => 0 for binary, 1 for ascii (other values raise an error)

### err = Save(points, normals)
points => the verticies of the triangles
normals => the normals of the triangles

