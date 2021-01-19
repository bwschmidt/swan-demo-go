owid = function() {
    "use-strict";

    // Parses a base 64 encoded string into a OWID tree.
    function parse(v) {

        function readByte(b) {
            return b.array[b.index++];
        }

        function readString(b) {
        var r = ""; 
        while (b.index < b.array.length && b.array[b.index] != 0) {
            r += String.fromCharCode(b.array[b.index++]);
        }
        b.index++;
        return r;
        }

        function readUint32(b) {
            return b.array[b.index++] | 
                b.array[b.index++]<<8 | 
                b.array[b.index++]<<16 |
                b.array[b.index++]<<24;
        }

        function readByteArray(b) {
            var c = readUint32(b);
            var r = b.array.slice(b.index, b.index + c)
            b.index += c;
            return r;
        }

        function readDate(b) {
            var h = readByte(b);
            var l = readByte(b);
            return {h, l};
        }

        function readOWID(b) {
            var o = Object();
            o.version = readByte(b);
            o.domain = readString(b);
            o.date = readDate(b);
            o.payload = readByteArray(b);
            o.signature = readByteArray(b);
            o.count = readUint32(b)
            o.children = []
            return o
        }

        // Decode the base64 string into a byte array.
        var b = Object();
        b.index = 0;
        b.array = Uint8Array.from(atob(v), c => c.charCodeAt(0));

        // Unpack the byte array into the OWID tree.
        var q = [];
        var r = readOWID(b);
        q.push(r);
        while (q.length > 0) {
            var n = q.shift();
            for (var i = 0; i < n.count; i++) {
                var c = readOWID(b);
                n.children.push(c)
                c.parent = n
                q.push(c)
            }
        }

        return r;
    }

    function getRoot(o) {
        var p = o;
        while (p) {
            o = p;
            p = p.parent;
        }
        return o;
    }

    function getOWIDFromIndex(t, i) {
        var p = t;
        i.forEach(v => {
            p = p.children[v];
        })
        return p;
    }

    this.appendName = function(e, s) {
        fetch("//" + parse(s).domain + "/owid/api/v1/creator")
            .then(r => r.json())
            .then(o => {
                var t = document.createTextNode(o.name);
                e.appendChild(t);
            }).catch(x => {
                console.log(x);
                var t = document.createTextNode(u);
                e.appendChild(t);
            });
    }

    this.appendAuditMark = function(e, r, t) {

        function importRsaKey(pem) {
            
            // Remove the header, footer and line breaks to get the PEM content.
            var lines = pem.split('\n');
            var pemContents = '';
            for (var i = 0; i < lines.length; i++) {
                if (lines[i].trim().length > 0 &&
                    lines[i].indexOf('-----BEGIN RSA PUBLIC KEY-----') < 0 &&
                    lines[i].indexOf('-----END RSA PUBLIC KEY-----') < 0) {
                    pemContents += lines[i].trim();
                }
            }

            // Import the public key with the SHA-256 hash algorithm.
            return window.crypto.subtle.importKey(
                "spki",
                Uint8Array.from(atob(pemContents), c => c.charCodeAt(0)),
                {
                name: "RSASSA-PKCS1-v1_5",
                hash: "SHA-256"
                },
                false,
                ["verify"]
            );
        }

        // Return the payload and the two date bytes as a byte array.
        function hashData(o) {
            var a = new Uint8Array(o.payload.length + 2);
            a.set(o.payload);
            a[a.length - 2] = o.date.h;
            a[a.length - 1] = o.date.l;
            return a;
        }

        // Append a failure HTML element.
        function returnFailed(e) {
            var t = document.createTextNode("Failed");
            e.appendChild(t);
        }

        // Append an audit HTML element.
        function addAuditMark(e, r) {
            var t = document.createElement("img");
            if (r) {
                t.src = "green.svg";
            } else {
                t.src = "red.svg";
            }
            e.appendChild(t);
        }

        // Use the well known end point for the alleged OWID creator. 
        function verifyOWIDWithAPI(r, t) {
            var o = parse(t)
            return fetch("//" + o.domain +
                "/owid/api/v" + o.version + "/verify" +
                "?root=" + encodeURIComponent(r) + 
                "&owid=" + encodeURIComponent(t),
                { method: "GET", mode: "cors", cache: "no-cache" })
                .then(r => r.json())
                .then(r => r.valid);
        }

        // Verify the payload of this OWID is the signature of the parent
        // OWID.
        function verifyOWIDWithPublicKey(r, t) {
            var o = parse(t)
            var a = Uint8Array.from(atob(r), c => c.charCodeAt(0));
            var b = Uint8Array.from(atob(t), c => c.charCodeAt(0));
            var m = new Int8Array(a.length + b.length);
            m.set(a);
            m.set(b, a.length);
            return fetch("//" + o.domain + "/owid/api/v1/creator",
                { mode: "cors", cache: "default" })
                .then(response => response.json())
                .then(c => importRsaKey(c.publicKeySPKI))
                .then(k => window.crypto.subtle.verify(
                    "RSASSA-PKCS1-v1_5",
                    k,
                    o.signature,
                    m));
        }

        // Valid the OWID against the creators public key OR if crypto not 
        // supported the well known end point for OWID creators.
        // if (window.crypto.subtle) {
        //     verifyOWIDWithPublicKey(r, t)
        //         .then(v => addAuditMark(e, v[0] && v[1]))
        //         .catch(x => {
        //             console.log(x);
        //             returnFailed(e);
        //         })
        // } else {
            verifyOWIDWithAPI(r, t)
                .then(r => addAuditMark(e, r))
                .catch(x => {
                    console.log(x);
                    returnFailed(e);
                });
        // }
    }
}