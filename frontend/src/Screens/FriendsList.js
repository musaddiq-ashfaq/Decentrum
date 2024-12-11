import React, { useEffect, useState } from "react";

const API_BASE_URL = "http://localhost:8081";

const FriendsList = () => {
  const [friends, setFriends] = useState([]);
  const [currentUser, setCurrentUser] = useState(null);

  useEffect(() => {
    const user = JSON.parse(localStorage.getItem("user"));
    setCurrentUser(user);
    if (user) {
      fetch(`${API_BASE_URL}/friends/${user.publicKey}`)
        .then((response) => {
          if (!response.ok) {
            throw new Error("Failed to fetch friends");
          }
          return response.json();
        })
        .then((data) => setFriends(data))
        .catch((error) => console.error("Error fetching friends:", error));
    }
  }, []);

  return (
    <div>
      <h2>Your Friends</h2>
      {friends.length > 0 ? (
        friends.map((friend, index) => (
          <div key={index}>
            <p>
              <strong>Name:</strong> {friend.name}
            </p>
            {/* <p>
              <strong>ID:</strong> {friend.id}
            </p> */}
          </div>
        ))
      ) : (
        <p>You have no friends.</p>
      )}
    </div>
  );
};

export default FriendsList;
