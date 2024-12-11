import { useEffect, useState } from "react";

const API_BASE_URL = "http://localhost:8081";

const FriendRequests = () => {
  const [friendRequests, setFriendRequests] = useState([]);
  const [currentUser, setCurrentUser] = useState(null);

  useEffect(() => {
    // Retrieve the current user from localStorage
    try {
      const userDataString = localStorage.getItem("user");
      if (userDataString) {
        const userData = JSON.parse(userDataString);
        setCurrentUser(userData);
        console.log("Retrieved current user data:", userData);
      } else {
        setCurrentUser(null);
      }
    } catch (error) {
      console.error("Error parsing localStorage data:", error);
      setCurrentUser(null);
      localStorage.clear(); // Clear localStorage if invalid data is found
    }
  }, []);

  useEffect(() => {
    if (currentUser && currentUser.publicKey) {
      console.log("Fetching friend requests for:", currentUser.publicKey);
      fetch(`${API_BASE_URL}/friend-requests/${currentUser.publicKey}`)
        .then((response) => {
          if (!response.ok) {
            throw new Error(
              `Error fetching friend requests: ${response.statusText}`
            );
          }
          return response.json();
        })
        .then((data) => {
          console.log("Fetched friend requests:", data);
          setFriendRequests(data);
        })
        .catch((error) =>
          console.error("Error fetching friend requests:", error)
        );
    }
  }, [currentUser]); // Add currentUser as a dependency

  const handleResponse = async (requestKey, response) => {
    console.log("in the func handle response");
    console.log(requestKey);
    console.log(response);
  
    // Find the friend request by requestKey to get the sender and receiver
    const friendRequest = friendRequests.find((req) => req.requestKey === requestKey);
  
    if (!friendRequest) {
      console.error("Friend request not found");
      return;
    }
  
    try {
      const res = await fetch(`${API_BASE_URL}/friend-request/respond`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          senderPublicKey: friendRequest.sender,
          receiverPublicKey: friendRequest.receiver,
          response,
        }),
      });
  
      if (res.ok) {
        // Remove the handled request from the state
        setFriendRequests(
          friendRequests.filter((req) => req.requestKey !== requestKey)
        );
      } else {
        console.error("Error responding to friend request:", await res.text());
      }
    } catch (error) {
      console.error("Error responding to friend request:", error);
    }
  };
  
  
  return (
    <div>
      <h2>Friend Requests</h2>
      {friendRequests.length === 0 && <p>No friend requests at the moment.</p>}
      {friendRequests.map((req) => (
        <div key={req.requestKey}>
          <span>{req.senderName || "Unknown Sender"}</span>
          <button onClick={() => handleResponse(req.requestKey, "accepted")}>
            Accept
          </button>
          <button onClick={() => handleResponse(req.requestKey, "rejected")}>
            Reject
          </button>
        </div>
      ))}
    </div>
  );
};

export default FriendRequests;
