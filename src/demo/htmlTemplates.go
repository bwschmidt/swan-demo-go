/* ****************************************************************************
 * Copyright 2020 51 Degrees Mobile Experts Limited (51degrees.com)
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 * use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
 * ***************************************************************************/

package demo

import "html/template"

var marTemplate = newHTMLTemplate("mar", `
<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
    <title>{{ .Title }}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="icon" href="data:;base64,=">
    <style>
        body {
            margin: 0;
			font-family: Verdana,sans-serif;
        }
        header {
		    background-color: {{ .BackgroundColor }};
            line-height: 70px;
            vertical-align: middle;
            border-bottom: solid black 2px;
            padding: 5px 20px;
            margin: 0;
            position: sticky;
            top: 0;
        }
        header h1 {
            display: inline;
        }
        main {
            margin: 0 auto;
            max-width: 57rem;
        }
        main section {
            margin: 4em auto;
            display: block;
        }
        main section h3, main section p {
            padding: 0 1em;
        }
        main section h3, main section p, main section pre, main section div {
            margin: 1em;
        }
        main section ul {
            list-style: none;
        }
        main section ul li {
            margin: 1em 0;
        }
        main section pre {
            background-color: lightgray;
            white-space: break-spaces;
            word-break: break-all;
            font-family: monospace;
        }
        main section pre {
            display: inline-block;
            padding: 0.5em;
        }
        footer {
            padding: 5px 20px;
            position: sticky;
            bottom: 0;
            left: 0;
            right: 0;
            background-color: white;
            border-top: solid black 2px;
        }
        footer ul {
            list-style: none;
            padding: 0;
        }
        footer ul li {
            display: inline;
        }
    </style>
    <script>
        {{ if ne .JSON "" }}
        var preferences = {{ .JSON }};
        console.log(preferences);
        {{ end }}
        // Get bid from the URL
        let urlParams = new URLSearchParams(window.location.search);
        let bid = urlParams.get('bid');

        // Parse the base64 string into JSON and parse to a JS Object.
        let data = JSON.parse(atob(bid));
        
        // Get the URLs to verify the OWIDs
        let urls = []
        for(let property in data['ids']) {
            let host = data['ids'][property].host;
            let owid = data['ids'][property].owid;
            if(host !== '' && owid != '') {
                urls.push('//' + host + '/owid/api/v1/decode-and-verify?owid=' +owid);
            }
        }
    
        // Verify that the IDs are valid
        Promise.all(urls.map(u=>fetch(u))).then(responses =>
            Promise.all(responses.map(res => res.json()))
        ).then(data => {
            data.forEach(d => console.log("'" + d.signature + "' is valid: " + d.valid))
        });
    </script>
</head>
<body>
    <header>
        <h1>{{ .Title }}</h1>
    </header>
    <main>
    <section>
        <h3>Find out more about our great products.</h3>
        <ul>
            <li>Item 1</li>
            <li>Item 2</li>
            <li>Item 3</li>
        </ul>
    </section>   
    {{ if .JSON }}
    <section>
        <h3>Incoming request in JSON</h3>
        <pre id="bid-json"></pre>
        <p>The incoming JSON can be verified using the following JavaScript</p>
        <pre>// Get bid from the URL
let urlParams = new URLSearchParams(window.location.search);
let bid = urlParams.get('bid');

// Parse the base64 string into JSON and parse to a JS Object.
let data = JSON.parse(atob(bid));

// Get the URLs to verify the OWIDs
let urls = []
for(let property in data['ids']) {
    let host = data['ids'][property].host;
    let owid = data['ids'][property].owid;
    if(host !== '' && owid != '') {
        urls.push('//' + host + '/owid/api/v1/decode-and-verify?owid=' +owid);
    }
}

// Verify that the IDs are valid
Promise.all(urls.map(u=>fetch(u))).then(responses =>
    Promise.all(responses.map(res => res.json()))
).then(data => {
    console.log(data)
});</pre>
    </section>
    {{ end }}    
    </main>
    <footer>
        <ul>
            {{ range $val := .Results }}
            <li>{{ $val.Key }}: {{ $val.Value }} | </li>
            {{ end }}
            <li><a href="{{ .SWANURL }}">Privacy Preferences</a></li>
        </ul>
    </footer>
    <script>
        document.getElementById("bid-json").innerText = atob(bid);
    </script>
</body>
</html>`)

var pubTemplate = newHTMLTemplate("pub", `
<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
    <title>Publisher | {{ .Title }}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="icon" href="data:;base64,=">
    <style>
        body {
            margin: 0;
			font-family: Verdana,sans-serif;
        }
        header {
		    background-color: {{ .BackgroundColor }};
            line-height: 70px;
            vertical-align: middle;
            padding: 5px 20px;
            margin: 0 auto;
            max-width: 57rem;
        }
        header h1 {
            display: inline;
            font-size: 1.5em;
        }
        main {
            margin: 0 auto;
            max-width: 57rem;
        }
        main section {
            margin: 4em auto;
            display: block;
        }
        main section h3, main section p, main section pre {
            margin: 1em;
        }
        main section pre, main section .inline-span {
            background-color: lightgray;
            white-space: break-spaces;
            word-break: break-all;
            font-family: monospace;
        }
        main section pre {
            display: inline-block;
            padding: 0.5em;
        }
        main section .inline-span {
            display: inline;
        }
        main section .button {
            border-radius: 0.5em;
            background-color: lightblue;
            padding: 0.5em;
            border: lightgrey solid 1px;
            text-decoration: none;
            margin: 1em auto;
            display: inline-block;
        }
        main section ul {
            list-style: none;
        }
        main section ul li {
            margin: 1em 0;
        }
        footer {
            position: sticky;
            bottom: 0;
            left: 0;
            right: 0;
            background-color: white;
            border-top: solid black 2px;
            word-break: break-all;
        }
        footer ul {
            list-style: none;
            text-align: center;
            padding: 0.5em;
            margin: 0;
        }
        footer ul li {
            display: inline;
        }
        main section .logos {
            list-style: none;
            display: flex;
            padding: 0;
            align-items: center;
            flex-flow: row wrap;
            justify-content: center;
        }
        main section .logos li {
            display: block;
            float: left;
            margin: 1em;
        }
        main section .logos li img {
            width: 96px;
        }
        main section div.container {
            padding: 1%;
            position: relative;
            display: table;
        }
        main section div.container .iconDetails {
            float: left;
            margin-right: 10px;
        }
        main section div.container .container-desc{
            display: table-cell;
            vertical-align: middle;
            padding-left: 10px;
        }
        main section div.container .text{
            margin: 0;
            padding: 0;
        }
        main section div.container .container-desc h4 {
            margin: 0px;
        }
        main section .slot {
            position: relative;
            font-family: Arial;
            width: 100%;
            max-width: 389px;
        }
        main section .slot img {
            width: 100%;
        }
        main section .slot .tooltip {
            position: absolute;
            top: 20px;
            right: 20px;
            background-color: royalblue;
            color: white;
            border-radius: 50%;
            padding: 8px;
            padding-left: 15px;
            padding-right: 15px;
        }
        main section .slot .tooltip .tooltiptext {
            visibility: hidden;
            width: 240px;
            background-color: black;
            color: #fff;
            text-align: center;
            border-radius: 6px;
            padding: 5px 0;
            position: absolute;
            z-index: 1;
            top: -5px;
            right: 105%;
        }
        main section .slot .tooltip:hover .tooltiptext {
            visibility: visible;
        }
        main section .label {
            color: darkgrey;
            font-size: 0.8em;
            text-align: center;
        }
    </style>
    <script>
        {{ if ne .JSON "" }}
        var preferences = {{ .JSON }};
        console.log(preferences);
        {{ end }}
        function verify(e, u) {
            fetch(u)
                .then(response => response.json())
                .then(data => {
                    if (data.valid == false) {
                        e.style.backgroundColor = "red";
                        e.innerText = "Bad";
                    } else {
                        e.style.backgroundColor = "green";
                        e.innerText = "Good";
                    }
                }).catch(() => e.style.backgroundColor = "red");
        }
        function creator(e, u) {
            fetch(u)
                .then(response => response.json())
                .then(data => {
                    e.innerText = data["name"];
                });
        }        
        function publicKey(e, u) {
            fetch(u)
                .then(response => response.json())
                .then(data => {
                    e.innerText = data["public-key"];
                });
        }    
        function text(e, u) {
            fetch(u)
                .then(response => response.text())
                .then(data => {
                    e.innerText = data;
                });
        }
        let url = "//{{ .Title }}/ssp/bid?" + 
            "cbid=" + encodeURIComponent("{{ .CBID.Value }}") +
            "&sid=" + encodeURIComponent("{{ .SID.Value }}") +
            "&oid=" + encodeURIComponent("{{ .OID }}") +
            "&allow=" + encodeURIComponent("{{ .Allow.Value }}")
        fetch(url)
            .then(response => response.json())
            .then(data => {
                let bid = btoa(JSON.stringify(data));
                let img = document.createElement("img");
                img.setAttribute("src", data["bid"]["creativeURL"]);
                img.setAttribute("width", "389px;");
                let link = document.createElement('a');
                link.setAttribute('href', data["bid"]["clickURL"] + "?bid=" + bid);
                link.appendChild(img)
                document.getElementById("slot1").appendChild(link);
        
                let urls = []
        
                for(let property in data['ids']) {
                    let host = data['ids'][property].host;
                    let owid = data['ids'][property].owid;
                    if(host !== '' && owid != '') {
                        urls.push('//' + host + '/owid/api/v1/decode-and-verify?owid=' + encodeURIComponent(owid));
                    }
                }
            
                Promise.all(urls.map(u=>fetch(u))).then(responses =>
                    Promise.all(responses.map(res => res.json()))
                ).then(data => {
                    let span1 = document.createElement("span")
                    span1.textContent="i";
                    let span2 = document.createElement("span")
                    span2.setAttribute('class', "tooltiptext");
                    let orgs = [...new Set(data.map(d => d.name))]
                    span2.textContent="The following companies were involved in supplying this advert: " + orgs.join(', ');
                    let tooltip = document.createElement("div")
                    tooltip.appendChild(span1)
                    tooltip.appendChild(span2)
                    tooltip.setAttribute('class', 'tooltip')
                    document.getElementById("slot1").appendChild(tooltip);
                });
            });
    </script>
</head>
<body>
    <header>
        <h1>Welcome to {{ .Title }}, powered by SWAN</h1>
    </header>
    <main>
    <section>
        <div id="slot1" class="slot">
        <div class="label">advertisement</div>
        </div>
    </section>
    <section>
        <h3>What is SWAN?</h3>
        <p>Shared Web Accountable Network (SWAN) is a secure, privacy-by-design ID that adds accountability to the Open Web. By enabling us to set your temporary SWAN ID, we and other SWAN supporters promise to respect your privacy choices. The SWAN network is a privacy-by-design method of enchancing people's cross-publisher experiences.</p>
        <ul>
        <li><strong>People</strong>: enhanced transparency, persistent personalization and privacy choices, while honoring people’s right to be forgotten.</li>
        <li><strong>Publishers</strong>: effective engagement, optimized advertising yield and accountable auditing to detect misappropriation.</li>
        <li><strong>Marketers</strong>: optimize cross publisher effectiveness, ensure you get what you pay for.</li>
        </ul>
        <p>SWAN was designed to provide enhanced transparency around data collection and enable stronger accountability and enforcement for those that violate your privacy. Recently browsers owned by the largest US publishers announced they intend to interfere with how we and other smaller publishers operate our business.</p>
        <p>Our partners support the competitive open web, which relies on the use of a fair, transparent, and privacy-centric identifier. We believe you deserve not only transparency and control, but an auditable view into which organizations were involved in displaying content to you on this publisher.</p>
        <p>To provide you access to our website, marketers fund our operations with advertising. In exchange, they need to measure and optimize their cross-publisher advertising as easily as they can within the Walled Gardens. However, marketers do not need to know your offline identity and SWAN members agree to keep your offline identity distinct from your digital activity.</p>
        <p>Like the World Wide Web, SWAN is a free, public service, operated by an open market of organizations that do not want or have a central controller. To ensure there is no single point of failure, there isn't a single SWAN domain, but many of them. To speed up your online experience, each browser can remember your privacy preferences, which reduces the number of times publishers need to ask you for your information.</p>
        <p>To learn more about the SWAN project, please visit our Open Source code repository <a href="https://github.com/51degrees/swan">here</a>.</p>
    </section>
    <section>
        <h3>SWAN supporters</h3>
        <ul class="logos">
            <li><img src="//51degrees.com/img/logo.png"></li>
            <li><img src="//zetaglobal.com/wp-content/uploads/2017/12/Top_Logo@2x.png"></li>
            <li><img src="//www.liveintent.com/assets/img/brand-assets/LiveIntentLogo-Horiz-Orange.png"></li>
        </ul>
    </section>
    {{ if .CBID }}
    <section>
        <h3>Common Browser ID (CBID)</h3>
        <p>SWAN provides a Common Browser ID that you can easily reset at any time. Here's the SWAN CBID for this browser.<p>
        <pre>{{ .CBID.AsOWID.PayloadAsString }}</pre>
        <p>You can reset this ID by clicking the reset button: [reset]</p>
        <p>SWAN secures your ID to ensure you can have an accountable audit log. Here's the secured version:<p>
        <pre>{{ .CBID.Value }}</pre>
        <p>Anyone can confirm that this ID was created by <span class="inline-span"><script>creator(document.scripts[document.scripts.length - 1].parentNode, '{{ .CBID.CreatorURL }}');</script></span> using this link.</p>
        <pre>{{ .CBID.VerifyURL }}</pre>
        <p>Go on. Tap the following button to check it's good.</p>
        <p><a class="button" onclick="verify(this, '{{ .CBID.VerifyURL }}')">Verify</a></p>
        <p>This shows that the domain <span class="inline-span">{{ .CBID.AsOWID.Domain }}</span> generated this ID on <span class="inline-span">{{ .CBID.AsOWID.Date }}</span>.</p>
        <p>The domain <span class="inline-span">{{ .CBID.AsOWID.Domain }}</span> used the following signature.</p>
        <pre>{{ .CBID.AsOWID.Signature }}</pre>
        <p>Because your online experience matters this publisher uses their public signing key so anyone can verify in microseconds.</p>
        <pre><script>publicKey(document.scripts[document.scripts.length - 1].parentNode, '{{ .CBID.CreatorURL }}');</script></pre>
    </section>
    {{ end }}
    {{ if .SID }}
    <section>
        <h3>Signed-in ID (SID)</h3>
        <p>If you wish to preserve your preferences across multiple browsers or devices, you can use SWAN to share your signed-in ID (SID). This relies on hashing a validated email you provide to register at this site. Here's the SID SWAN generated from whatever you entered.</p>
        <pre>{{ .SID.AsOWID.PayloadAsPrintable }}</pre>
        <p>Just like CBID it's secured to make it verifiable. Here's the longer version.</p>
        <pre>{{ .SID.Value }}</pre>
        <p>When all of this is decoded and verified it looks like this.</p>
        <pre><script>text(document.scripts[document.scripts.length - 1].parentNode, '{{ .SID.DecodeAndVerifyURL }}');</script></pre>
        <p>SID and CBID are all implemented in SWAN using the Open Web ID schema. It's open source and you can find out more <a href="https://github.com/51degrees/owid">here</a>.</p>
    </section>
    {{ end }}
    {{ if .OID }}
    <section>
        <h3>Offer ID (OID)</h3>
        <p>Here's an advertising OfferID generated for this page request.</p>
        <pre>{{ .OID }}</pre>
        <p>When this is decoded it looks like this.</p>
        <pre>{{ .UnpackOID }}</pre>
    </section>
    {{ end }}
    {{ if .Allow }}
    <section>
        <h3>Preferences</h3>
        <p>Responsible addressable marketing requires both respecting your preferences and providing you transparency as to which organizations were involved in delivering you personalized content. SWAN has recorded your personalization preferences as.</p>
        <pre>{{ .Allow.AsOWID.PayloadAsString }}</pre>
        <p>Just like your Common Browser ID, we secure your preferences too. Your preference token is:</p>
        <pre>{{ .Allow.Value }}</pre>
        <p>You can change your preferences any time by clicking the My preferences button.</p>
        <p><a class="button" href="?privacy=update">My preferences</a></p>
        <p>If you want to only temporarily change your preference, you can using a new incognito or private browsing tab.</p>
    </section>
    {{ end }}
    <section>
        <h3>Find out more about the open source projects used in this demo.</h3>
        <div class='container'>
            <div>
                <img class='iconDetails' src='https://github.com/51Degrees/swift/raw/main/images/swift_128px_72dpi_v2.png'>
            </div>	
                <div class='container-desc'>
                <h4><a href="https://github.com/51degrees/swift">SWIFT</a></h4>
                <p class="text">Shared Web InFormaTion (SWIFT) is a browser-agnostic method to share information across web domains.</p>
            </div>
            </div>
            <div class='container'>
            <div>
                <img class='iconDetails' src='https://github.com/51Degrees/owid/raw/main/images/owl_128px_72dpi.png'>
            </div>	
                <div class='container-desc'>
                <h4><a href="https://github.com/51degrees/owid">OWID</a></h4>
                <p class="text">Open Web ID (OWID) is a privacy-by-design schema for ID.</p>
            </div>
            </div>
            <div class='container'>
            <div>
                <img class='iconDetails' src='https://github.com/51Degrees/swan/raw/main/images/swan_128px_72dpi.png'>
            </div>	
                <div class='container-desc'>
                <h4><a href="https://github.com/51degrees/swan">SWAN</a></h4>
                <p class="text">Shared Web Accountable Network (SWAN) brings it all together to support digital marketing use cases.</p>
            </div>
        </div>
    </section>
    <section>
        <h3>Visit these other domains</h3>
        <p>Just visit these other domains that are part of SWAN to see how the same data is shared across multiple domains.</p>
        <ul>
            {{ range $val := .Pubs }}
            <li><a href="//{{ $val }}">{{ $val }}</a></li>
            {{ end }}
        </ul>
    </section>
    </main>
    <footer>
        <ul>
            <li>CBID:{{ .CBID.AsOWID.PayloadAsString }}</li>
            <li>SID:{{ .SID.AsOWID.PayloadAsPrintable }}</li>
            <li>Pref.:{{ .Allow.AsOWID.PayloadAsString }}</li>
            <li><a href="?privacy=update">Privacy Preferences</a></li>
        </ul>
    </footer>
</body>
</html>`)

func newHTMLTemplate(n string, h string) *template.Template {
	return template.Must(template.New(n).Parse(h))
}
