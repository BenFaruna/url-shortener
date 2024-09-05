async function submitForm(e) {
    e.preventDefault()
    console.log(e.srcElement[0])
    const urlInput = document.querySelector("[name=url]")
    const data = {
        "method": "POST",
        "body": JSON.stringify({ "url": urlInput.value }),
        "headers": {
            "Content-Type": "application/json"
        }
    }
    const res = await fetch("/api/v1/shorten", data)
    console.log(await res.json())
    location.reload()
}