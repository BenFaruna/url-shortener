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
    const res = await fetch("/api/v1/address/shorten", data)
    if (!res.ok) {
        return
    }
    await res.json()
    location.reload()
}

async function deleteUrl(e, urlId, urlShort) {
    e.preventDefault()
    console.log(urlId, e)
    const data = {
        "method": "DELETE",
        "body": JSON.stringify({ "url-id": urlId }),
        "headers": {
            "Content-Type": "application/json"
        }
    }
    const res = await fetch(`/api/v1/address/${urlShort}`, data)
    console.log(await res.json())
    // await res.json()
    location.reload()
}