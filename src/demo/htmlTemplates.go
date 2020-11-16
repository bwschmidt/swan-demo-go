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
            display: flex;
            flex-wrap: wrap;
        }
        main section {
            flex: 0 0 33.3333%;
        }
        main section h3, main section p {
            padding: 0 1em;
        }
        main section ul {
            list-style: none;
        }
        main section ul li {
            margin: 1em 0;
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

        </ul>
    </section>           
    </main>
    <footer>
        <ul>
            {{ range $val := .Results }}
            <li>{{ $val.Key }} : {{ $val.Value }} | </li>
            {{ end }}
            <li><a href="{{ .SWANURL }}">Privacy Preferences</a></li>
        </ul>
    </footer>   
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
        main section h3, main section p, main section pre, main section div {
            margin: 1em;
        }
        main section pre, main section span {
            background-color: lightgray;
            white-space: break-spaces;
            word-break: break-all;
            font-family: monospace;
        }
        main section pre {
            display: block;
            padding: 0.5em;
        }
        main section span {
            display: inline;
        }
        main section div {
            padding: 0.5em;
        }
        main section div .button {
            border-radius: 0.5em;
            background-color: lightblue;
            padding: 0.5em;
            border: lightgrey solid 1px;
        }
        main section ul {
            list-style: none;
        }
        main section ul li {
            margin: 1em 0;
        }
        main section img {
            width: 128px;
        }
        footer {
            position: sticky;
            bottom: 0;
            left: 0;
            right: 0;
            background-color: white;
            border-top: solid black 2px;
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
    </script>
</head>
<body>
    <header>
        <h1>Publisher: {{ .Title }}</h1>
    </header>
    <main>
    <section>
        <h3>Welcome to SWAN : the future of open web advertising</h3>
        <p>You're experiencing a privacy compliant (GDPR and CCPA), secure, cross domain, shared consent and personalization solution that works on every web browser today and in the future.</p>
        <ul>
        <li><strong>People</strong> : enhacned transparency, persistent choices (no more consent fatigue) and "right to be forgotten".</li>
        <li><strong>Publishers</strong> : effective engagement, optimized advertising yield and no misappropriation.</li>
        <li><strong>Marketers</strong> : optimize cross publisher effectiveness, ensure you get what you pay for.</li>
        <li><strong>Decentralized</strong> : open source with multiple implementors forming a network.</li> 
        </ul>
        <p>Read on for a brief introduction to SWAN. To find out more go <a href="https://github.com/51degrees/swan">here</a>.</p>
    </section>
    <section>
        <h3>SWAN supporters</h3>
        <ul>
            <li><img src="//51degrees.com/img/logo.png"></li>
            <li><img src="//zetaglobal.com/wp-content/uploads/2017/12/Top_Logo@2x.png"></li>
            <li><img src="//www.liveintent.com/assets/img/brand-assets/LiveIntentLogo-Horiz-Orange.png"></li>
        </ul>
    </section>
    {{ if .CBID }}
    <section>
        <h3>Common Browser ID (CBID)</h3>
        <p>SWAN provides a Common Browser ID that can easily be reset and is verifiable. Here's the SWAN CBID for this browser.<p>
        <pre>{{ .CBID.AsOWID.PayloadAsString }}</pre>
        <p>It's wrapped up with other data to make it verifiable supporting accountability requirements. Here's the full version.<p>
        <pre>{{ .CBID.Value }}</pre>
        <p>Anyone can confirm it was created by <span><script>creator(document.scripts[document.scripts.length - 1].parentNode, '{{ .CBID.CreatorURL }}');</script></span> operating the domain <span>{{ .CBID.AsOWID.Domain }}</span> on <span>{{ .CBID.AsOWID.Date }}</span>.</p>
        <p>They embedded the following cryptographic signature.</p>
        <pre>{{ .CBID.AsOWID.Signature }}</pre>
        <p>They provide a URL to verify the CBID came from them.</p>
        <pre>{{ .CBID.VerifyURL }}</pre>
        <p>Go on. Tap the following button to check it's good.</p>
        <div><a class="button" onclick="verify(this, '{{ .CBID.VerifyURL }}')">Verify</a></div>
        <p>When speed matters they publish their public signing key so anyone can verify in microseconds.</p>
        <pre><script>publicKey(document.scripts[document.scripts.length - 1].parentNode, '{{ .CBID.CreatorURL }}');</script></pre>
    </section>
    {{ end }}
    {{ if .UUID }}
    <section>
        <h3>Universally Unique ID (UUID)</h3>
        <p>SWAN supports cross device IDs such as hashed email addresses. Here's the UUID SWAN generated from whatever you entered.</p>
        <pre>{{ .UUID.AsOWID.PayloadAsString }}</pre>
        <p>Just like CBID it's wrapped up with other data to make it verifiable. Here's the longer version.</p>
        <pre>{{ .UUID.Value }}</pre>
        <p>When all of this is decoded and verified it looks like this.</p>
        <pre><script>text(document.scripts[document.scripts.length - 1].parentNode, '{{ .UUID.DecodeAndVerifyURL }}');</script></pre>
        <p>UUIDs and CBID are all implemented in SWAN using the Open Web ID schema. It's open source and your can find out more <a href="https://github.com/51degrees/owid">here</a>.</p>
    </section>
    {{ end }}
    {{ if .Allow }}
    <section>
        <h3>Preferences</h3>
        <p>Responsible digital marketing is also about capturing and respecting people's preferences. SWAN has recorded your personalization preferences as.</p>
        <pre>{{ .Allow.AsOWID.PayloadAsString }}</pre>
        <p>Just like IDs they're also wrapped up with other data to make them auditable.</p>
        <pre>{{ .Allow.Value }}</pre>
        <p>Preferences can be changed at any time <a href="{{ .SWANURL }}">here</a>.</p>
        <p>Consent fatigue is eliminated because people set their preferences once for all web sites that support SWAN. People get to publisher's content faster knowing their privacy preferences are respected. Marketers get a supply chain they can audit and trust.</p>
    </section>
    {{ end }}
    <section>
        <h3>Decentralized</h3>
        <p>Because SWAN is operated by multiple hosts working together to form a network there is no central operator.</p>
        <p>There isn't a single SWAN domain, there are many of them, and they change all the time. Block lists won't work for those seeking to interfere with publishers and marketers digital businesses.</p>
        <p>To speed things up every browser is assigned a home domain. This reduces the number of times multiple SWAN domains needed to be accessed to store or retrieve information.</p>
        <p>This does mean that publishers and marketers will need to redirect initial visits to the SWAN home domain. Most of the time the browser will be directed straight back again in a fraction of a second with the CBID, UUID and preferences. SWAN is quicker than current privacy preferences dialogue boxes.</p>
        <p>SWAN is built on the Shared Web InFormaTion (SWIFT) open source project for sharing data cross domain in a post third party cookie world. Find our more about SWIFT <a href="https://github.com/51degrees/swift">here</a>.</p>
    </section>
    <section>
        <h3>The future</h3>
        <p>We think people will really like the transparency and control that SWAN provides. If they don't want personalization then they just opt out once, and don't have to do it for every web site over and over again. If they're happy with providing limited personal data in exchange for free content and services funded by advertising then they can express that choice once and don't need to be nagged over and over again.</p>
        <p>As SWAN expands hosts will be able to produce browser extensions to improve the performance of the network. As not everyone will want or know how to install extensions there will also be a need for a web based solution.</p>
    </section>
    <section>
        <h3>Find out more about the open source projects used in this demo.</h3>
        <ul>
            <li><a href="https://github.com/51degrees/swift">SWIFT</a> used to shared data across multiple domains and all web browsers.</li>
            <li><a href="https://github.com/51degrees/owid">OWID</a> auditable IDs.</li>
            <li><a href="https://github.com/51degrees/swan">SWAN</a> bringing it all together to support digital marketing use cases.</li>
        </ul>
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
            {{ range $val := .Results }}
            <li>{{ $val.Key }} : {{ $val.AsOWID.PayloadAsString }}</li>
            {{ end }}
            <li><a href="{{ .SWANURL }}">Privacy Preferences</a></li>
        </ul>
    </footer>
</body>
</html>`)

func newHTMLTemplate(n string, h string) *template.Template {
	return template.Must(template.New(n).Parse(h))
}
