import numpy as np
import matplotlib.pyplot as plt
import pdb

def read_jagged_matrix(path):
	data = []
	with open(path, 'r') as f:
		for line in f.readlines():
			data.append([float(d) for d in line.strip().split(' ')])
	return data	
    
def make_jagged_square(data, fake_number):
    col_len = [len(d) for d in data]
    max_col_len = np.max(col_len)
    mat = np.zeros((len(data), max_col_len))
    for row in range(mat.shape[0]):
        start_copy = max_col_len - col_len[row]
        for col in range(mat.shape[1]):
            mat[row,col] = fake_number
            if col > start_copy:
                mat[row,col] = data[row][col - start_copy]
    return mat
def make_both_sided(mat):
    mat = np.concatenate((mat, np.flip(mat, axis=1)), axis=1)
    return np.flip(mat, axis=0)
    
# --- MAIN --- #
side = 'deck'
# mz
jmz = read_jagged_matrix('out/'+side+'_mz.txt')
mz = make_both_sided(make_jagged_square(jmz, -1.0))

# angle dev
jdza = read_jagged_matrix('out/'+side+'_dz_angle.txt')  
dza = make_both_sided(make_jagged_square(jdza, -1.0))

# height dev
jdzh = read_jagged_matrix('out/'+side+'_dz_height.txt')
dzh = make_both_sided(make_jagged_square(jdzh, -1.0))

# feedrate
jfr = read_jagged_matrix('out/'+side+'_feedrate.txt')
fr = make_both_sided(make_jagged_square(jfr, 500))

fig = plt.figure()

plt.subplot(131)
plt.title('Z height deviation')
x1 = plt.imshow(dzh)
fig.colorbar(x1)

plt.subplot(132)
plt.title('Z angle deviation')
x2 = plt.imshow(dza)
fig.colorbar(x2)

plt.subplot(133)
plt.title('Feedrate')
x3 = plt.imshow(fr)
fig.colorbar(x3)

plt.show()
#pdb.set_trace()
