const API_BASE_URL = "http://localhost:8081";

export const fetchUsers = async () => {
  const response = await fetch(`${API_BASE_URL}/users`);
  if (!response.ok) throw new Error("Failed to fetch users");
  return response.json();
};

export const fetchFriendRequests = async (userPublicKey) => {
  const response = await fetch(`${API_BASE_URL}/friend-requests/${userPublicKey}`);
  if (!response.ok) throw new Error("Failed to fetch friend requests");
  return response.json();
};

export const sendFriendRequest = async (sender, receiver) => {
  const response = await fetch(`${API_BASE_URL}/friend-request/send`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ senderPublicKey: sender, receiverPublicKey: receiver }),
  });
  if (!response.ok) throw new Error("Failed to send friend request");
  return response.json();
};

export const respondToFriendRequest = async (sender, receiver, responseStatus) => {
  const response = await fetch(`${API_BASE_URL}/friend-request/respond`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      senderPublicKey: sender,
      receiverPublicKey: receiver,
      response: responseStatus,
    }),
  });
  if (!response.ok) throw new Error("Failed to respond to friend request");
  return response.json();
};
