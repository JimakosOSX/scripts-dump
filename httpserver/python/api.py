# api.py

from netifaces import interfaces, ifaddresses, AF_INET
from datetime import datetime
from time import time


import psutil
import distro
import socket
import shutil
import flask
import json
import os


app = flask.Flask(__name__)
application = app

@app.route("/")
def show_info():
    # Track execution time
    start_time = time()
    # FQDN
    with open("/etc/hostname") as file:
        fqdn_pretty = file.read().rstrip()

    # CPU usage
    cpu_perc = str(psutil.cpu_percent()) + '%'
    # RAM usage
    ram = str( psutil.virtual_memory()[2] ) + '% used'
    # Distribution
    distro_pretty = distro.name(pretty=True)
    # HDD Usage
    total, used, free = shutil.disk_usage("/")
    used_gb = (used // (2**30))
    total_gb = (total // (2**30))
    hdd = str(used_gb) + 'GiB / ' + str(total_gb) + 'GiB'
    # Load average
    values = os.getloadavg()
    loads = []

    for item in values:
        loads.append(round(item, 2))

    loadavg_pretty = str(loads)[1:-1]

    # IP addresses
    ip_list = []
    for interface in interfaces():
        for link in ifaddresses(interface)[AF_INET]:
            ip_list.append(str(link['addr']))

    ip_list_pretty = ' '.join(repr(item) for item in ip_list)
    ip_list_prettier = ip_list_pretty.replace("'", '')

    # Date
    now = datetime.now()
    dt_string = now.strftime("%d/%m/%Y %H:%M")
    time_pretty = dt_string

    # Execution time calculation
    execution_time = (time() - start_time)
    execution_time_rounded = round(execution_time, 3)
    execution_time_str = str(execution_time_rounded) + " s"

    # User Agent
    user_agent = flask.request.headers.get('User-Agent')

    # SimpleX-Chat fingerprint for easier access
    # https://github.com/simplex-chat/simplex-chat/blob/stable/docs/SERVER.md

    with open("/etc/opt/simplex/fingerprint") as file:
        fingerprint = file.read().rstrip()

    smp_server_address = "smp://" + fingerprint + ":PASSWORD@" + fqdn_pretty

    data = [
        { "FQDN": fqdn_pretty },
        { "CPU": cpu_perc },
        { "RAM": ram },
        { "DISTRO": distro_pretty },
        { "HDD": hdd },
        { "LOAD": loadavg_pretty },
        { "IP": ip_list_prettier },
        { "EXEC_TIME": execution_time_str },
        { "USER_AGENT": user_agent },
        { "DATE": time_pretty },
        { "SMP_ADDRESS": smp_server_address },
    ]

    return flask.jsonify(data)
    #return flask.render_template('index.html', json_data=data)

