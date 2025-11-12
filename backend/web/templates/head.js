(function () {
  const headRoot = document.getElementById("head-root");
  if (!headRoot) {
    return;
  }

  headRoot.innerHTML = `
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">

    <!-- Base -->
    <meta name="description" content="Safex — Safe encrypted secrets exchange">
    <meta name="author" content="Gonfff">
    <meta name="description"
        content="Share sensitive information securely with self-destructing messages protected by PIN codes. Free, open-source, and privacy-focused.{{end}}">
    <meta name="keywords"
        content="password sharing, file sharing, secure messaging, encrypted messages, self-destructing messages, PIN protection, secret sharing, sensitive information, privacy, security, p2p, open-source">

    <!-- Canonical URL -->
    <link rel="canonical" href="{{.URL}}">

    <!-- Open Graph -->
    <meta property="og:type" content="website">
    <meta property="og:url" content="{{.URL}}">
    <meta property="og:title" content="Safe exchange">
    <meta property="og:description" content="Safex — Safe encrypted secrets exchange">
    <meta property="og:image" content="{{.BaseURL}}/static/img/logo.png">
    <meta property="og:image:width" content="630">
    <meta property="og:image:height" content="630">

    <!-- Style -->
    <link rel='shortcut icon' href='/static/img/favicon.ico' type='image/x-icon'>
    <link rel="icon" type="image/png" sizes="32x32" href="/static/img/favicon-32x32.png">
    <link rel="icon" type="image/png" sizes="16x16" href="/static/img/favicon-16x16.png">
    <meta name="theme-color" content="#1a1a1a">
    <!-- <link rel="icon" href="/static/vendor/favicon.ico" type="image/x-icon"> -->
    <!-- <link href="/static/app.css" rel="stylesheet"> -->
    <!-- <script src="https://unpkg.com/htmx.org@1.9.12" defer></script> -->
    <link href="./output.css" rel="stylesheet">

    <title>Safe exchange</title>
</head>`;
})();
