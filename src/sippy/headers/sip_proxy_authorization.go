// Copyright (c) 2003-2005 Maxim Sobolev. All rights reserved.
// Copyright (c) 2006-2015 Sippy Software, Inc. All rights reserved.
// Copyright (c) 2015 Andrii Pylypenko. All rights reserved.
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
package sippy_header

type SipProxyAuthorization struct {
    *SipAuthorization
}

var _sip_proxy_authorization_name normalName = newNormalName("Proxy-Authorization")

func NewSipProxyAuthorizationWithBody(body *SipAuthorizationBody) *SipProxyAuthorization {
    super := NewSipAuthorizationWithBody(body)
    super.normalName = _sip_proxy_authorization_name
    return &SipProxyAuthorization{
        SipAuthorization : super,
    }
}

func NewSipProxyAuthorization(realm, nonce, uri, username, algorithm string) *SipProxyAuthorization {
    super := NewSipAuthorization(realm, nonce, uri, username, algorithm)
    super.normalName = _sip_proxy_authorization_name
    return &SipProxyAuthorization{
        SipAuthorization : super,
    }
}

func CreateSipProxyAuthorization(body string) []SipHeader {
    super := createSipAuthorizationObj(body)
    super.normalName = _sip_proxy_authorization_name
    return []SipHeader{ &SipProxyAuthorization{
            SipAuthorization : super,
        },
    }
}
