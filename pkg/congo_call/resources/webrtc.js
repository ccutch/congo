(function () {

  let remoteStream;

  let peer = new RTCPeerConnection({
    iceServers: [
      {
        urls: ['stun:stun1.l.google.com:19302', 'stun:stun2.l.google.com:19302'],
      },
    ],
    iceCandidatePoolSize: 10,
  });

  window.setupLocalDevice = function () {
    remoteStream = new MediaStream();
    navigator.mediaDevices.getUserMedia({ video: true, audio: true })
      .then(setupLocalStream)
      .then(handlePeerStreams);

    function setupLocalStream(stream) {
      stream.getTracks().forEach(track => peer.addTrack(track, stream));
      peer.createOffer().then(setupLocalDescription);
      document.getElementById("local-video").srcObject = stream;
    }

    function handlePeerStreams() {
      peer.ontrack = function (event) {
        event.streams[0].getTracks().forEach(track => {
          remoteStream.addTrack(track, event.streams[0])
        });
      }
    }
  };

  window.createCall = function (nameInput) {
    fetch("/call", {
      headers: { "Content-Type": "application/json" },
      method: "POST",
      body: JSON.stringify({ name: nameInput.value }),
    })
      .then(function (response) { return response.json() })
      .then(listenForCandidates)
      .then(createCallOffer)
      .then(listenForEvents);

    function listenForCandidates(room) {
      peer.onicecandidate = function (event) {
        if (event.candidate) {
          const data = new FormData()
          data.append("candidate", JSON.stringify(event.candidate));
          fetch("/_call/" + room.ID, {
            headers: { "Content-Type": "application/json" },
            method: "PUT",
            body: data,
          });
        }
      };
      return room;
    }

    function createCallOffer(room) {
      peer.createOffer().then(function (offer) {
        console.log("created-offer", offer);
        peer.setLocalDescription(new RTCSessionDescription(offer));

        const data = new FormData()
        data.append("offer", offer.type)
        data.append("sdp", offer.sdp)

        fetch("/_call/" + room.ID, {
          headers: { "Content-Type": "application/json" },
          method: "PUT",
          body: data,
        });
      });
      return room;
    }
  }

  window.joinCall = function (roomID) {
    fetch("/call/" + roomID)
      .then(function (response) { return response.json() })
      .then(listenForCandidates)
      .then(handleRemotePeer)
      .then(answerCallOffer)
      .then(listenForEvents)

    function listenForCandidates(room) {
      peer.onicecandidate = function (event) {
        if (event.candidate) {
          const data = new FormData()
          data.append("candidate", JSON.stringify(event.candidate));
          fetch("/_call/" + room.ID, {
            headers: { "Content-Type": "application/json" },
            method: "PUT",
            body: data,
          });
        }
      };
      return room;
    }

    function handleRemotePeer(room) {
      const offer = new RTCSessionDescription(room.Offer);
      return peer.setRemoteDescription(offer);
    }

    function answerCallOffer(room) {
      return peer.createAnswer().then(function (answer) {
        console.log("created-answer", answer);
        peer.setLocalDescription(new RTCSessionDescription(answer));

        const data = new FormData()
        data.append("answer", answer.type)
        data.append("sdp", answer.sdp)

        fetch("/_call/" + room.ID, {
          headers: { "Content-Type": "application/json" },
          method: "PUT",
          body: data,
        });

        return room;
      });
    }
  };

  function listenForEvents(room) {
    const stream = new EventSource("/_call/" + room.ID + "/events");
    stream.onmessage = function (event) {
      if (event.data.startsWith("call-answer")) {

        console.log("call-answer", event.data);
        const answer = JSON.parse(event.data);
        peer.setRemoteDescription(new RTCSessionDescription(answer));

      } else if (event.data.startsWith("candidate")) {

        console.log("candidate", event.data);
        const candidate = JSON.parse(event.data);
        peer.addIceCandidate(new RTCIceCandidate(candidate));

      } else
        console.log("unknown", event.type, event.data);
    };
  }

})();