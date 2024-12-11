
import { useEffect, useState } from "react";
import { fetchUsers, sendFriendRequest } from "../utils/api";
import { Card, CardContent, CardHeader, CardTitle } from "../Components/ui/card";
import { Button } from "../Components/ui/button";

const UserList = ({ friendRequests }) => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(false);
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
    const loadUsers = async () => {
      try {
        const users = await fetchUsers();
        setUsers(users);
      } catch (error) {
        console.error("Failed to load users", error);
      }
    };
    loadUsers();
  }, []);

  const handleSendRequest = async (receiverPublicKey) => {
    console.log("inside the handle send req fun")
    setLoading(true);
    try {
      await sendFriendRequest(currentUser.publicKey, receiverPublicKey);
      alert("Friend request sent!");
    } catch (error) {
      console.error("Failed to send friend request", error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card className="w-full max-w-md mx-auto">
      <CardHeader>
        <CardTitle>All Users</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {users
            .filter((user) => user.publicKey !== currentUser?.publicKey)
            .map((user) => (
              <div
                key={user.publicKey}
                className="flex items-center justify-between p-2 border rounded"
              >
                {/* Display user name */}
                <span>{user.name}</span>
                {/* Button for sending friend request */}
                <Button onClick={() => handleSendRequest(user.publicKey)} disabled={loading}>
                  {loading ? "Sending..." : "Add Friend"}
                </Button>
              </div>
            ))}
        </div>
      </CardContent>
    </Card>
  );
};

export default UserList;