# Smart Gateway Setup

Notes about how to setup the Smart Gateway components from client to server.

## virthost

### Firewall

    firewall-cmd --add-port=20001/tcp
    firewall-cmd --add-port=5672/tcp

### QDR configuration

    cd ~
    mkdir ~/qpid-dispatch
    cat > qpid-dispatch/qdrouterd.conf <<EOF

    # See the qdrouterd.conf (5) manual page for information about this
    # file's format and options.

    router {
        mode: interior
        id: CENTRAL_ROUTER.virthost

    }

    # This is for _client_ connections (senders and receivers, collectd wirtes to this )
    # to connect on port 5672:
    listener {
        host: 0.0.0.0
        port: amqp
        authenticatePeer: no
        saslMechanisms: ANONYMOUS
    }

    listener {
        role: inter-router
        host: 192.168.122.1
        port: 20001
        authenticatePeer: no
        saslMechanisms: ANONYMOUS
    }


    # This establishes an outgoing inter-router connection to QPD.B
    # listener
    #
    #connector {
    #    role: inter-router
    #    host: Either an IP address (IPv4 or IPv6) or hostname on which the router should connect
    #    port: 20002
    #    saslMechanisms: ANONYMOUS
    #}


    # Various address prefix -> distribution pattern
    # configurations:
    #
    address {
        prefix: closest
        distribution: closest
    }

    address {
        prefix: multicast
        distribution: multicast
    }

    address {
        prefix: unicast
        distribution: closest
    }

    address {
        prefix: exclusive
        distribution: closest
    }

    address {
        prefix: broadcast
        distribution: multicast
    }
    EOF


### Start QDR

    docker run -it --volume=`pwd`/qpid-dispatch/:/etc/qpid-dispatch.conf.d/ --net=host nfvpe/qpid-dispatch-router --config=/etc/qpid-dispatch.conf.d/qdrouterd.conf


## cloud-node-1

### QDR configuration

    cd ~
    mkdir qpid-dispatch
    cat > qpid-dispatch/qdrouterd.conf <<EOF

    # See the qdrouterd.conf (5) manual page for information about this
    # file's format and options.

    router {
        mode: interior
        id: BAROMETER_ROUTER.cloud-node-1

    }

    # This is for _client_ connections (senders and receivers, collectd wirtes to this )
    # to connect on port 5672:
    listener {
        host: 0.0.0.0
        port: amqp
        authenticatePeer: no
        saslMechanisms: ANONYMOUS
    }


    # This establishes an outgoing inter-router connection to QPD.B
    # listener
    #
    connector {
        role: inter-router
        host: 192.168.122.1
        port: 20001
        saslMechanisms: ANONYMOUS
    }


    # Various address prefix -> distribution pattern
    # configurations:
    #
    address {
        prefix: closest
        distribution: closest
    }

    address {
        prefix: multicast
        distribution: multicast
    }

    address {
        prefix: unicast
        distribution: closest
    }

    address {
        prefix: exclusive
        distribution: closest
    }

    address {
        prefix: broadcast
        distribution: multicast
    }

### Start QDR

    docker run -it --publish 172.17.0.1:5672:5672/tcp --volume `pwd`/qpid-dispatch/:/etc/qpid-dispatch.conf.d/:Z nfvpe/qpid-dispatch-router --config=/etc/qpid-dispatch.conf.d/qdrouterd.conf

### barometer-collectd configuration

    cd ~
    mkdir collect_config
    cat > collect_config/barometer.conf <<EOF
    LoadPlugin amqp1
    <Plugin amqp1>
    <Transport "name">
        Host "172.17.0.1"
        Port "5672"
    #    User "guest"
    #    Password "guest"
        Address "collectd"
    #    <Instance "log">
    #        Format JSON
    #        PreSettle false
    #    </Instance>
        <Instance "notify">
            Format JSON
            PreSettle true
            Notify true
        </Instance>
        <Instance "telemetry">
            Format JSON
            PreSettle false
        </Instance>
    </Transport>
    </Plugin>
    EOF

### Start barometer-collectd

    docker run -ti --net=host -v `pwd`/collect_config:/opt/collectd/etc/collectd.conf.d -v /var/run:/var/run -v /tmp:/tmp --privileged nfvpe/barometer-collectd