import { useEffect, useState } from "react";
import { Button } from "../Components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../Components/ui/card";
import { Input } from "../Components/ui/input";

const API_BASE_URL = "http://localhost:8081"; // Base backend URL

const ChatApp = () => {
  const [currentUser, setCurrentUser] = useState(null); // Store current user
  const [users, setUsers] = useState([]);
  const [selectedUser, setSelectedUser] = useState(null);
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState("");
  const [loading, setLoading] = useState({
    users: false,
    messages: false,
    sending: false,
  });

  // Fetch current user from local storage using the new approach
  useEffect(() => {
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

  const fetchUsers = async () => {
    setLoading((prev) => ({ ...prev, users: true }));
    try {
      const response = await fetch(`${API_BASE_URL}/users`);
      const userData = await response.json();

      const formattedUsers = userData.map((user) => ({
        username: user.name,
        ...user,
      }));

      setUsers(formattedUsers);

      if (formattedUsers.length > 0 && !selectedUser) {
        setSelectedUser(formattedUsers[0]);
      }
    } catch (err) {
      console.error("Failed to fetch users:", err);
    } finally {
      setLoading((prev) => ({ ...prev, users: false }));
    }
  };

  const fetchMessages = async () => {
    if (!currentUser || !selectedUser) return;

    setLoading((prev) => ({ ...prev, messages: true }));
    try {
      const response = await fetch(`${API_BASE_URL}/chat`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          operation: "get",
          username: selectedUser.username,
          senderUsername: currentUser.name,
        }),
      });

      const messageData = await response.json();
      console.log("API Response for messages:", messageData); 
    
      const normalizedMessages = messageData.map((msg) => ({
        senderUsername: msg.username || currentUser.name,
        messages: msg,
        timestamp: msg.timestamp || new Date().toISOString(),
      }));

      console.log("Fetched messages:", normalizedMessages);
      setMessages(normalizedMessages);
    } catch (err) {
      console.error("Failed to fetch messages:", err);
    } finally {
      setLoading((prev) => ({ ...prev, messages: false }));
    }
  };

  const sendMessage = async () => {
    if (!newMessage.trim()) return;

    setLoading((prev) => ({ ...prev, sending: true }));
    try {
      const response = await fetch(`${API_BASE_URL}/chat`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          operation: "send",
          username: currentUser.name,
          receiverUsername: selectedUser.username,
          plainText: newMessage,
        }),
      });

      if (!response.ok) {
        throw new Error("Failed to send message");
      }

      const newMessageObj = {
        senderUsername: currentUser.name,
        plainText: newMessage,
        timestamp: new Date().toISOString(),
      };

      setMessages((prevMessages) => [...prevMessages, newMessageObj]);
      setNewMessage("");

      await fetchMessages();
    } catch (err) {
      console.error("Failed to send message:", err);
    } finally {
      setLoading((prev) => ({ ...prev, sending: false }));
    }
  };

  useEffect(() => {
    if (currentUser) fetchUsers();
  }, [currentUser]);

  useEffect(() => {
    let intervalId;
    if (selectedUser) {
      fetchMessages();
      intervalId = setInterval(fetchMessages, 5000);
    }
    return () => intervalId && clearInterval(intervalId);
  }, [selectedUser]);

  return (
    <div className="flex h-screen">
      {/* Sidebar for available users */}
      <div className="w-1/4 bg-gray-100 p-4 overflow-y-auto">
        <h2 className="text-xl font-bold mb-4">Available Users</h2>
        {users
          .filter((user) => user.username !== currentUser?.name)
          .map((user) => (
            <div
              key={user.username}
              className={`p-2 mb-2 cursor-pointer rounded ${
                selectedUser?.username === user.username
                  ? "bg-blue-500 text-white"
                  : "hover:bg-gray-200"
              }`}
              onClick={() => {
                console.log("Selected user:", user);
                setSelectedUser(user);
              }}
            >
              {user.username}
            </div>
          ))}
      </div>

      {/* Chat window */}
      <div className="w-3/4 flex flex-col">
        {selectedUser ? (
          <Card className="h-full flex flex-col">
            <CardHeader>
              <CardTitle>
                Chat with {selectedUser.username}
                {loading.messages && (
                  <span className="ml-2 text-sm text-gray-500">
                    (Loading messages...)
                  </span>
                )}
              </CardTitle>
            </CardHeader>
            <CardContent className="flex-grow overflow-y-auto flex flex-col">
              {messages.map((msg, index) => (
                <div
                  key={index}
                  className={`mb-2 p-2 rounded max-w-xs ${
                    msg.senderUsername === currentUser.name
                      ? "bg-blue-100 self-end text-right"
                      : "bg-gray-100 self-start text-left"
                  }`}
                >
                  <p className="text-sm text-gray-700">
                    {msg.senderUsername !== currentUser.name && (
                      <span className="font-semibold">{msg.senderUsername}: </span>
                    )}
                    {msg.plainText || msg.messages}
                  </p>
                  <span className="block text-xs text-gray-500 mt-1">
                    {new Date(msg.timestamp).toLocaleString()}
                  </span>
                </div>
              ))}
            </CardContent>
            <div className="flex p-4">
              <Input
                value={newMessage}
                onChange={(e) => setNewMessage(e.target.value)}
                placeholder="Type a message..."
                className="flex-grow mr-2"
                disabled={loading.sending}
              />
              <Button
                onClick={sendMessage}
                disabled={loading.sending || !newMessage.trim()}
              >
                {loading.sending ? "Sending..." : "Send"}
              </Button>
            </div>
          </Card>
        ) : (
          <div className="flex items-center justify-center h-full text-gray-500">
            Select a user to start chatting
          </div>
        )}
      </div>
    </div>
  );
};

export default ChatApp;
