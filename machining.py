import json
import numpy as np
import serial
import time
import pdb

"""
=============================
====  PRIVATE FUNCTIONS  ====
=============================
"""
def _confirmingReturnMessage(tg):
    isConfirmed = False
    message = tg.readlines()
    for line in message:
        line_split = line.split('"')
        if (len(line_split)>1 and line_split[1]=='r' ):
            isConfirmed = True
    return isConfirmed

def _setValue(tg, config, value):
    print('Sets: "'+str(config)+'" to '+str(value))
    tg.write('{"'+str(config)+'":'+str(value)+'}\n')
    time.sleep(0.1)
    return tg.readlines()
def _setGcodeValue(tg, config):
    print('Sets: "'+str(config)+'"')
    tg.write('{"gc":"'+str(config)+'"}\n')
    time.sleep(0.1)
    return tg.readlines()

def _readValue(tg, config):
    tg.write('{'+str(config)+':null}\n')
    time.sleep(0.1)
    reply = tg.readline().rstrip()
    dict_reply = json.loads(reply)
    value = dict_reply['r'][config]
    print('Reads value: "'+str(config)+'" as '+ str(value))
    return value

def _move(tg, x, y, z, f):
    tg.write('{"gc":"g1 x+'+str(x)+' y'+str(y)+' z'+str(z)+' f'+str(f)+'"}\n')
    _printStatusReport(tg)
    time.sleep(0.1)
    _setGcodeValue(tg,'M2')
    time.sleep(0.1)
    return tg.readline()

def _move_x(tg, x, f):
    tg.write('{"gc":"g1 x+'+str(x)+' f'+str(f)+'"}\n')
    _printStatusReport(tg)
    time.sleep(0.05)
    return tg.readline()
    
def _move_y(tg, y, f):
    tg.write('{"gc":"g1 y+'+str(y)+' f'+str(f)+'"}\n')
    _printStatusReport(tg)
    time.sleep(0.05)
    return tg.readline()

def _move_z(tg, z, f):
    tg.write('{"gc":"g1 z+'+str(z)+' f'+str(f)+'"}\n')
    _printStatusReport(tg)
    time.sleep(0.05)
    return tg.readline()

def _disableAmaxLimitSwitch(tg):
    _setValue(tg, 'asx', 0)
    value = _readValue(tg, 'asx')   
    if value == 0:
        print('Success!')
    if value != 0:
        print('Panic wrong value set!')

def _enableAmaxLimitSwitch(tg):
    _setValue(tg, 'asx', 2)
    value = _readValue(tg, 'asx')
    if value == 2:
        print('Success!')
    if value != 2:
        print('Panic wrong value set!')

def _setGcodeValues(tg):
    _setGcodeValue(tg, "G40")
    _setGcodeValue(tg, "G49")
    _setGcodeValue(tg, "G80")
    _setGcodeValue(tg, "G54")
    _setGcodeValue(tg, "G90")
    _setGcodeValue(tg, "G21")
    _setGcodeValue(tg, "M3 S1000")

def _setGcodeMovesFromFile(tg, fname):
    tg.flush()
    with open(fname) as f:
        lines = f.readlines()

    for i in range(4):
        tg.write(lines[i])
        print(lines[i].rstrip())
    index = 4
    len_lines = len(lines)
    while True:
        report = tg.readline()
        print(report.rstrip())
        if report.split('"')[1] == 'r':
            tg.flush()
            print(lines[index].rstrip())
            print('Adding line: ' + str(index) + ' of ' + str(len_lines) + ' ' + str(100*index/len_lines) + '%')
            tg.write(lines[index])
            index = index + 1
            if index == len(lines) - 1:
                break
    
def _connect():
    print('connecting...')
    tg = serial.Serial('/dev/ttyUSB0', baudrate=115200, timeout=2, rtscts=True, xonxoff=False)
    time.sleep(1)
    tg.write('$ej=1\n')
    time.sleep(0.2)
    lines = tg.readlines()
    for line in lines:
        print(line.rstrip())
    if len(lines)>0:
        print('...we\'re in\n')
    return tg

def _disconnect(tg):
    print('disconnecting...\n')
    tg.close()

def _printStatusReport(tg):
    tg.write('{"sr":"n"}\n')
    message = tg.readline()
    while message != '':
        print(message.rstrip())
        message = tg.readline()
        time.sleep(0.05)
    print

def _millControl(command):
    tg = _connect()
    if (command.split()[0]=='G1' or command.split()[0]=='g1'):
        val = _setGcodeValue(tg, command)
    else:
        config = command.split()[0]
        if len(command.split())>1:
            command = command.split()[1]
        else:
            command = 'n'
        val = _setValue(tg, config, command)
    _disconnect(tg)
    return val

def _getCurrentPosition():
    tg = _connect()
    x = _readValue(tg, 'posx')
    y = _readValue(tg, 'posy')
    z = _readValue(tg, 'posz')
    return [x, y, z]

def _hasDoneHoming():
    m = _millControl('home n')
    dm = json.loads(m[0])
    if dm['r']['home']:
        return True
    return False

def _getSoftLimits():
    tg = _connect()
    com = ['xtn', 'xtm', 'ytn', 'ytm', 'ztn', 'ztm']
    r_com = [_readValue(tg,c) for c in com]  
    _disconnect(tg)
    return r_com

def _isAtHomePosition():
    # getting the home position
    with open('settings.json') as f:
        home_pos = np.array(json.load(f)['HomingOffset'])
    # getting the current position
    current_pos = np.array(_getCurrentPosition())
    # testing
    if np.sum(current_pos - home_pos) == 0:
        return True
    return False
    
def _isAtStartPosition():
    # getting the home position
    with open('settings.json') as f:
        start_pos = np.array(json.load(f)['HomingOffset'])
    start_pos[1] = 1200
    # getting the current position
    current_pos = np.array(_getCurrentPosition())
    # testing
    if np.sum(current_pos - start_pos) == 0:
        return True
    return False

def _moveToStartPosition():
    # reads the homing offset
    with open('settings.json') as f:
        data = json.load(f)
    dx = data['HomingOffset'][0]
    dz = data['HomingOffset'][2]
    
    # move to start position
    tg = _connect()
    _disableAmaxLimitSwitch(tg)
    _move_z(tg, dz, 1000)
    _move_x(tg, dx, 1000)
    _move_y(tg, 1200, 2000)
    _enableAmaxLimitSwitch(tg)
    _disconnect(tg)

"""
=============================
====  PUBLIC FUNCTIONS  ====
=============================
"""
# --- moves --- #
def homing():
    # moving home
    tg = _connect()
    _disableAmaxLimitSwitch(tg)
    tg.write('{"gc":"g28.2 x0 y0 z0"}')
    _printStatusReport(tg)
    
    # reads the setting.json file
    with open('settings.json') as f:
        data = json.load(f)
    
    # sets the homing offset
    homingOffset = data['HomingOffset']
    dx = homingOffset[0]
    dy = homingOffset[1]
    dz = homingOffset[2]
    print('Homing Offset = ', dx, dy, dz)
    tg.write('{"gc":"g28.3 x'+str(dx)+' y'+str(dy)+' z'+str(dz)+'"}')
    _printStatusReport(tg)
    _disconnect(tg)
    
def     millDeck():
    # reads the homing offset
    with open('settings.json') as f:
        data = json.load(f)
    dx = data['HomingOffset'][0]
    dy = data['HomingOffset'][1]
    dz = data['HomingOffset'][2]
    
    tg = _connect()
    # move to start position
    _disableAmaxLimitSwitch(tg)
    _move_z(tg, dz, 2000)
    _move_x(tg, dx, 2000)
    _move_y(tg, 1200, 2000)
    _enableAmaxLimitSwitch(tg)
    # mill deck
    _setGcodeValues(tg)
    _setGcodeMovesFromFile(tg, 'cam/deck.gc')
    _setGcodeValue(tg,'M2')
    # move to start position
    _move_z(tg, dz, 2000)
    _move_x(tg, dx, 2000)
    _move_y(tg, 1200, 2000)
    
    # move to home position 
    _move_z(tg, dz, 2000)
    _move_x(tg, dx, 2000)
    _move_y(tg, dy, 2000)
    
    _disconnect(tg)

def millBottom():
    # reads the homing offset
    with open('settings.json') as f:
        data = json.load(f)
    dx = data['HomingOffset'][0]
    dy = data['HomingOffset'][1]
    dz = data['HomingOffset'][2]
    
    tg = _connect()
    # move to start position
    _disableAmaxLimitSwitch(tg)
    _move_z(tg, dz, 1000)
    _move_x(tg, dx, 1000)
    _move_y(tg, 1200, 2000)
    _enableAmaxLimitSwitch(tg)
    # mill deck
    _setGcodeValues(tg)
    _setGcodeMovesFromFile(tg, 'cam/bottom.gc')
    _setGcodeValue(tg,'M2')
    # move to start position
    _move_z(tg, dz, 1500)
    _move_x(tg, dx, 1500)
    _move_y(tg, 1200, 1500)
    
    # move to home position 
    _move_z(tg, dz, 1500)
    _move_x(tg, dx, 1500)
    _move_y(tg, dy, 1500)
    
    _disconnect(tg)


