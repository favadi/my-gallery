const submitImageLike = async (event) => {
    const data = {image_id: Number(event.target.getAttribute("image-id"))};
    const response = await fetch(`/likes`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    })

    if (response.status === 200) {
        event.target.parentNode.classList.add("liked");
    }
}

document.querySelectorAll(".like-button")
    .forEach(elem => elem.addEventListener("click", submitImageLike));
