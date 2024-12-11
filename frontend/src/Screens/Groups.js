import { useEffect, useState } from "react";
import { Button } from "../Components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../Components/ui/card";
import { Input } from "../Components/ui/input";

const API_BASE_URL = "http://localhost:8081"; // Base backend URL

const CreateGroup = () => {
  const [currentUser, setCurrentUser] = useState(null);
  const [users, setUsers] = useState([]); // Add state for users
  const [groupName, setGroupName] = useState("");
  const [selectedUsers, setSelectedUsers] = useState([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    // Retrieve the current user from localStorage
    try {
      const userDataString = localStorage.getItem("user");
      if (userDataString) {
        const userData = JSON.parse(userDataString);
        setCurrentUser(userData);
        console.log("Retrieved user data:", userData);
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
    // Fetch the list of users
    const fetchUsers = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/users`);
        if (!response.ok) {
          throw new Error("Failed to fetch users");
        }
        const data = await response.json();
        console.log("Retrieved users data:", data); 
        setUsers(data); // Populate the users state
      } catch (error) {
        console.error("Error fetching users:", error);
      }
    };

    fetchUsers();
  }, []);

  const toggleUserSelection = (username) => {
    setSelectedUsers((prev) =>
      prev.includes(username)
        ? prev.filter((user) => user !== username)
        : [...prev, username]
    );
  };

  const createGroup = async () => {
    if (!groupName.trim() || selectedUsers.length === 0) {
      alert("Please provide a group name and select at least one user.");
      return;
    }

    setLoading(true);
    try {
      const response = await fetch(`${API_BASE_URL}/groups`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name:groupName,
          members: [currentUser.name, ...selectedUsers],
        }),
      });
     
      if (!response.ok) {
        throw new Error("Failed to create group");
      }

      alert("Group created successfully!");
      setGroupName("");
      setSelectedUsers([]);
    } catch (error) {
      console.error("Error creating group:", error);
      alert("An error occurred while creating the group.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card className="w-full max-w-md mx-auto">
      <CardHeader>
        <CardTitle>Create New Group</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="mb-4">
          <Input
            value={groupName}
            onChange={(e) => setGroupName(e.target.value)}
            placeholder="Enter group name"
            className="w-full"
            disabled={loading}
          />
        </div>
        <div className="mb-4">
          <h3 className="text-lg font-semibold mb-2">Select Users:</h3>
          <div className="max-h-40 overflow-y-auto">
            {users
              .filter((user) => user.name !== currentUser?.name)
              .map((user) => (
                <div key={user.name} className="flex items-center mb-2">
                  <input
                    type="checkbox"
                    id={`user-${user.name}`}
                    checked={selectedUsers.includes(user.name)}
                    onChange={() => toggleUserSelection(user.name)}
                    disabled={loading}
                    className="mr-2"
                  />
                  <label htmlFor={`user-${user.name}`}>
                    {user.name}
                  </label>
                </div>
              ))}
          </div>
        </div>
        <Button onClick={createGroup} disabled={loading || !groupName.trim()}>
          {loading ? "Creating..." : "Create Group"}
        </Button>
      </CardContent>
    </Card>
  );
};

export default CreateGroup;
