// Copyright (c) 2003-2005 Maxim Sobolev. All rights reserved.
// Copyright (c) 2006-2014 Sippy Software, Inc. All rights reserved.
// Copyright (c) 2016 Andriy Pylypenko. All rights reserved.
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
// list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation and/or
// other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
// ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
package sippy

import (
    "math/rand"
    "net"
    "time"
    "strings"

    "sippy/conf"
)

type Rtp_proxy_client_impl interface {
    GoOnline()
    GoOffline()
}

type Rtp_proxy_opts struct {
    No_version_check    *bool
    Spath               *string
    Nworkers            *int
    Bind_address        *sippy_conf.HostPort
}

func (self *Rtp_proxy_opts) no_version_check() bool {
    if self == nil || self.No_version_check == nil {
        return false
    }
    return *self.No_version_check
}

func (self *Rtp_proxy_opts) spath() *string {
    if self == nil {
        return nil
    }
    return self.Spath
}

func (self *Rtp_proxy_opts) bind_address() *sippy_conf.HostPort {
    if self == nil {
        return nil
    }
    return self.Bind_address
}

type Rtp_proxy_client_base struct {
    me              Rtp_proxy_client_impl
    transport       rtp_proxy_transport
    proxy_address   string
    online          bool
    sbind_supported bool
    tnot_supported  bool
    caps_done       bool
    shut_down       bool
    hrtb_retr_ival  time.Duration
}

type rtp_proxy_transport interface {
    IsLocal() bool
    send_command(string, func(string))
}

func (self *Rtp_proxy_client_base) IsLocal() bool {
    return self.transport.IsLocal()
}

func (self *Rtp_proxy_client_base) IsOnline() bool {
    return self.online
}

func (self *Rtp_proxy_client_base) SBindSupported() bool {
    return self.sbind_supported
}

func (self *Rtp_proxy_client_base) TNotSupported() bool {
    return self.tnot_supported
}

func (self *Rtp_proxy_client_base) GetProxyAddress() string {
    return self.proxy_address
}

func randomize(x time.Duration, p float64) time.Duration {
    return time.Duration(float64(x) * (1.0 + p * (1.0 - 2.0 * rand.Float64())))
}

/*
CAPSTABLE = {"20071218":"copy_supported", "20080403":"stat_supported", \
  "20081224":"tnot_supported", "20090810":"sbind_supported", \
  "20150617":"wdnt_supported"}

class Rtpp_caps_checker(object):
    caps_requested = 0
    caps_received = 0
    rtpc = nil

    def __init__(self, rtpc):
        self.rtpc = rtpc
        rtpc.caps_done = false
        for vers in CAPSTABLE.iterkeys():
            self.caps_requested += 1
            rtpc.send_command("VF %s" % vers, self.caps_query_done, vers)

    def caps_query_done(self, result, vers):
        self.caps_received -= 1
        vname = CAPSTABLE[vers]
        if result == "1":
            setattr(self.rtpc, vname, true)
        else:
            setattr(self.rtpc, vname, false)
        if self.caps_received == 0:
            self.rtpc.caps_done = true
            self.rtpc = nil

class Rtp_proxy_client_base(Rtp_proxy_client_udp, Rtp_proxy_client_stream):
    worker = nil
    address = nil
    online = false
    copy_supported = false
    stat_supported = false
    tnot_supported = false
    sbind_supported = false
    wdnt_supported = false
    shut_down = false
    proxy_address = nil
    caps_done = false
    sessions_created = nil
    active_sessions = nil
    active_streams = nil
    preceived = nil
    ptransmitted = nil
    hrtb_ival = 1.0
    hrtb_retr_ival = 60.0
*/

func NewRtp_proxy_client_base(me Rtp_proxy_client_impl, global_config sippy_conf.Config, address net.Addr, opts *Rtp_proxy_opts) (*Rtp_proxy_client_base, error) {
    var err error
    var rtpp_class func(*Rtp_proxy_client_base, sippy_conf.Config, net.Addr, *Rtp_proxy_opts) (rtp_proxy_transport, error)
    self := &Rtp_proxy_client_base{
        me          : me,
        caps_done   : false,
        shut_down   : false,
        hrtb_retr_ival  : 60 * time.Second,
    }
    //print "Rtp_proxy_client_base", address
    if address == nil && opts.spath() != nil {
        var rtppa net.Addr
        a := *opts.spath()
        if strings.HasPrefix(a, "udp:") {
            tmp := strings.SplitN(a, ":", 3)
            if len(tmp) == 2 {
                rtppa, err = net.ResolveUDPAddr("udp", tmp[1] + ":22222")
            } else {
                rtppa, err = net.ResolveUDPAddr("udp", tmp[1] + ":" + tmp[2])
            }
            if err != nil { return nil, err }
            self.proxy_address, _, err = net.SplitHostPort(rtppa.String())
            if err != nil { return nil, err }
            rtpp_class = NewRtp_proxy_client_udp
        } else if strings.HasPrefix(a, "udp6:") {
            tmp := strings.SplitN(a, ":", 2)
            a := tmp[1]
            rtp_proxy_host, rtp_proxy_port := a, "22222"
            if a[len(a)-1] != ']' {
                idx := strings.LastIndexByte(a, ':')
                if idx < 0 {
                    rtp_proxy_host = a
                } else {
                    rtp_proxy_host, rtp_proxy_port = a[:idx], a[idx+1:]
                }
            }
            if rtp_proxy_host[0] != '[' {
                rtp_proxy_host = "[" + rtp_proxy_host + "]"
            }
            rtppa, err = net.ResolveUDPAddr("udp", rtp_proxy_host + ":" + rtp_proxy_port)
            if err != nil { return nil, err }
            self.proxy_address, _, err = net.SplitHostPort(rtppa.String())
            if err != nil { return nil, err }
            rtpp_class = NewRtp_proxy_client_udp
        } else if strings.HasPrefix(a, "tcp:") {
            tmp := strings.SplitN(a, ":", 3)
            if len(tmp) == 2 {
                rtppa, err = net.ResolveTCPAddr("tcp", tmp[1] + ":22222")
            } else {
                rtppa, err = net.ResolveTCPAddr("tcp", tmp[1] + ":" + tmp[2])
            }
            if err != nil { return nil, err }
            self.proxy_address, _, err = net.SplitHostPort(rtppa.String())
            if err != nil { return nil, err }
            rtpp_class = NewRtp_proxy_client_stream
        } else if strings.HasPrefix(a, "tcp6:") {
            tmp := strings.SplitN(a, ":", 2)
            a := tmp[1]
            rtp_proxy_host, rtp_proxy_port := a, "22222"
            if a[len(a)-1] != ']' {
                idx := strings.LastIndexByte(a, ':')
                if idx < 0 {
                    rtp_proxy_host = a
                } else {
                    rtp_proxy_host, rtp_proxy_port = a[:idx], a[idx+1:]
                }
            }
            if rtp_proxy_host[0] != '[' {
                rtp_proxy_host = "[" + rtp_proxy_host + "]"
            }
            rtppa, err = net.ResolveTCPAddr("tcp", rtp_proxy_host + ":" + rtp_proxy_port)
            if err != nil { return nil, err }
            self.proxy_address, _, err = net.SplitHostPort(rtppa.String())
            if err != nil { return nil, err }
            rtpp_class = NewRtp_proxy_client_stream
        } else {
            if strings.HasPrefix(a, "unix:") {
                rtppa, err = net.ResolveUnixAddr("unix", a[5:])
            } else if strings.HasPrefix(a, "cunix:") {
                rtppa, err = net.ResolveUnixAddr("unix", a[6:])
            } else {
                rtppa, err = net.ResolveUnixAddr("unix", a)
            }
            self.proxy_address = global_config.SipAddress().String()
            rtpp_class = NewRtp_proxy_client_stream
        }
        self.transport, err = rtpp_class(self, global_config, rtppa, opts)
        if err != nil {
            return nil, err
        }
    } else if strings.HasPrefix(address.Network(), "udp") {
        self.transport, err = NewRtp_proxy_client_udp(self, global_config, address, opts)
        if err != nil {
            return nil, err
        }
        self.proxy_address, _, err = net.SplitHostPort(address.String())
        if err != nil { return nil, err }
    } else {
        self.transport, err = NewRtp_proxy_client_stream(self, global_config, address, opts)
        if err != nil {
            return nil, err
        }
        self.proxy_address = global_config.SipAddress().String()
    }
    if ! opts.no_version_check() {
        self.version_check()
    } else {
        self.caps_done = true
        self.online = true
    }
    return self, nil
}

func (self *Rtp_proxy_client_base) SendCommand(cmd string, cb func(string)) {
    self.transport.send_command(cmd, cb)
}
/*
    def reconnect(self, *args, **kwargs):
        self.rtpp_class.reconnect(self, *args, **kwargs)
*/

func (self *Rtp_proxy_client_base) version_check() {
    if self.shut_down {
        return
    }
    self.transport.send_command("V", self.version_check_reply)
}

func (self *Rtp_proxy_client_base) version_check_reply(version string) {
    if self.shut_down {
        return
    }
    if version == "20040107" {
        self.me.GoOnline()
    } else if self.online {
        self.me.GoOffline()
    } else {
        t := NewTimeout(self.version_check, nil, randomize(self.hrtb_retr_ival, 0.1), 1, nil)
        t.Start()
    }
}
/*
    def heartbeat(self):
        //print "heartbeat", self, self.address
        if self.shut_down:
            return
        self.send_command("Ib", self.heartbeat_reply)

    def heartbeat_reply(self, stats):
        //print "heartbeat_reply", self.address, stats, self.online
        if self.shut_down:
            return
        if ! self.online:
            return
        if stats == nil:
            self.active_sessions = nil
            self.go_offline()
        else:
            sessions_created = active_sessions = active_streams = preceived = ptransmitted = 0
            for line in stats.splitlines():
                line_parts = line.split(":", 1)
                if line_parts[0] == "sessions created":
                    sessions_created = int(line_parts[1])
                elif line_parts[0] == "active sessions":
                    active_sessions = int(line_parts[1])
                elif line_parts[0] == "active streams":
                    active_streams = int(line_parts[1])
                elif line_parts[0] == "packets received":
                    preceived = int(line_parts[1])
                elif line_parts[0] == "packets transmitted":
                    ptransmitted = int(line_parts[1])
                self.update_active(active_sessions, sessions_created, active_streams, preceived, ptransmitted)
        Timeout(self.heartbeat, randomize(self.hrtb_ival, 0.1))
*/

func (self *Rtp_proxy_client_base) GoOnline() {
    if self.shut_down {
        return
    }
    if ! self.online {
        rtpp_cc = Rtpp_caps_checker(self)
        self.online = true
        self.heartbeat()
    }
}

func (self *Rtp_proxy_client_base) GoOffline() {
    if self.shut_down {
        return
    }
    //print "go_offline", self.address, self.online
    if self.online {
        self.online = false
        Timeout(self.version_check, randomize(self.hrtb_retr_ival, 0.1))
    }
}
/*
    def update_active(self, active_sessions, sessions_created, active_streams, preceived, ptransmitted):
        self.sessions_created = sessions_created
        self.active_sessions = active_sessions
        self.active_streams = active_streams
        self.preceived = preceived
        self.ptransmitted = ptransmitted

    def shutdown(self):
        if self.shut_down: // do not crash when shutdown() called twice
            return
        self.shut_down = true
        self.rtpp_class.shutdown(self)
        self.rtpp_class = nil

    def get_rtpc_delay(self):
        self.rtpp_class.get_rtpc_delay(self)
*/
