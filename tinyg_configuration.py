import machining as m
import pdb
import json, sys

def get_keys():
    with open('tinyg_configuration_keys.json') as f:
            data = json.load(f)
    return list(data.keys())

def write_conf_to_file(filename):
    keys = get_keys()
    keys.sort()
    vals = []
    tg = m._connect()
    for k in keys:
        print('reading: ', str(k))
        vals.append(m._readValue(tg, str(k)))
    m._disconnect(tg)
    mydict = {}
    for i in range(len(keys)):
        mydict[keys[i]] = vals[i]
    with open(filename,'w') as f:
        json.dump(mydict, f, sort_keys=True, indent=4, separators=(',', ': '))

#--- MAIN ---#
write_conf_to_file('test2.json')
"""
keys = get_keys()
keys.sort()
vals = []
tg = m._connect()
for k in keys:
    print('reading: ', str(k))
    vals.append(m._readValue(tg, str(k)))
m._disconnect(tg)
mydict = {}
for i in range(len(keys)):
    mydict[keys[i]] = vals[i]
with open('test.json','w') as f:
    json.dump(mydict, f, sort_keys=True, indent=4, separators=(',', ': '))
 print('done')
"""
