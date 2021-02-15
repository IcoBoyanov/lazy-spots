
const sofiaCenter = { lat: 42.6893643, lng: 23.3255209 };
let map = {}
let markers = {};

function initMap() {
    $("#map").empty();
    markers = {};
    map = new google.maps.Map(document.getElementById("map"), {
      zoom: 14,
      center: sofiaCenter,
    });
  }

  $("#laod").on("click", function() {
      loadStavaPlaces()
  });
let marker = {}
async function loadStavaPlaces() {
    data = await fetch("http://localhost:8888/places", {
        // mode: 'no-cors',
    })
    .then(response => response.text())
    .then(data => JSON.parse(data))
    .then(data => { return data } ) 
    .catch(error => { return { "Status": 500 } });

    for (let p of data.data) {
        new google.maps.Marker({
            position: p,
            map,
            title: "",
        });
    }

}
