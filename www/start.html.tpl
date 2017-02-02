<html>
    <head>
        <title>Ubuntu snapd device</title>
        <link rel="stylesheet" type="text/css" media="all" href="/css/ubuntu-styles.css"/>
    </head>
    <style>
        body {
            margin: 0;
            min-height: 100vh;
            color: #E95420;
            font-family: 'ubuntu';
        }
        div {
            position: absolute;
            text-align: center;
            left: 50%;
            top: 50%;
            width: 100%;
            transform: translateX(-50%) translateY(-50%);
        }
        .title {
            color: #2C001E;
            font-size: 4em;
        }
        .ip {
            font-size: 2em;
        }
    </style>
    <body>
        <div>
            <p class="title">Connect a device to one of:</p>
            {{range .Addrs}}
            <p class="ip">{{ . }}</p>
            {{end}}
        </div>
    </body>
</html>