import { useEffect, useState } from "react";
import { Button } from "../Components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../Components/ui/card";
import { Input } from "../Components/ui/input";
import { Send, User, MessageCircle, Loader } from 'lucide-react';
import Navbar from "./Navbar";

const API_BASE_URL = "http://localhost:8081";

const ChatApp = () => {
  const [currentUser, setCurrentUser] = useState(null);
  const [users, setUsers] = useState([]);
  const [selectedUser, setSelectedUser] = useState(null);
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState("");
  const [loading, setLoading] = useState({
    users: false,
    messages: false,
    sending: false,
  });

  useEffect(() => {
    try {
      const userDataString = localStorage.getItem("user");
      if (userDataString) {
        const userData = JSON.parse(userDataString);
        setCurrentUser(userData);
      } else {
        setCurrentUser(null);
      }
    } catch (error) {
      console.error("Error parsing localStorage data:", error);
      setCurrentUser(null);
      localStorage.clear();
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
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          operation: "get",
          username: selectedUser.username,
          senderUsername: currentUser.name,
        }),
      });
      const messageData = await response.json();
      const normalizedMessages = messageData.map((msg) => ({
        senderUsername: msg.username || currentUser.name,
        messages: msg,
        timestamp: msg.timestamp || new Date().toISOString(),
      }));
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
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          operation: "send",
          username: currentUser.name,
          receiverUsername: selectedUser.username,
          plainText: newMessage,
        }),
      });
      if (!response.ok) throw new Error("Failed to send message");
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
    <div className="bg-gradient-animation min-h-screen flex flex-col">
      <Navbar />
      <div className="flex-grow p-4 flex items-center justify-center">
        <Card className="w-full max-w-6xl bg-white/90 backdrop-blur-md shadow-xl rounded-xl overflow-hidden">
          <div className="flex h-[calc(80vh-64px)]">
            {/* User list sidebar */}
            <div className="w-1/3 border-r border-gray-200 overflow-y-auto">
              <CardHeader className="sticky top-0 bg-white z-10">
                <CardTitle className="text-2xl font-bold text-[#052a47] flex items-center">
                  <MessageCircle className="h-6 w-6 mr-2" />
                  Chats
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                {users
                  .filter((user) => user.username !== currentUser?.name)
                  .map((user) => (
                    <div
                      key={user.username}
                      className={`p-4 cursor-pointer transition-all duration-200 flex items-center ${
                        selectedUser?.username === user.username
                          ? "bg-[#4dbf38] text-white"
                          : "hover:bg-gray-100"
                      }`}
                      onClick={() => setSelectedUser(user)}
                    >
                      <User className="h-8 w-8 mr-3 p-1 bg-[#052a47] text-white rounded-full" />
                      <span className="font-medium">{user.username}</span>
                    </div>
                  ))}
              </CardContent>
            </div>

            {/* Chat window */}
            <div className="w-2/3 flex flex-col">
              {selectedUser ? (
                <>
                  <CardHeader className="border-b border-gray-200 bg-white sticky top-0 z-10">
                    <CardTitle className="text-xl font-semibold text-[#052a47] flex items-center">
                      <User className="h-8 w-8 mr-3 p-1 bg-[#4dbf38] text-white rounded-full" />
                      {selectedUser.username}
                      {loading.messages && (
                        <Loader className="ml-2 h-4 w-4 animate-spin text-[#4dbf38]" />
                      )}
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="flex-grow overflow-y-auto p-4 space-y-4">
                    {messages.map((msg, index) => (
                      <div
                        key={index}
                        className={`flex ${
                          msg.senderUsername === currentUser.name ? "justify-end" : "justify-start"
                        }`}
                      >
                        <div
                          className={`rounded-lg p-3 max-w-[70%] ${
                            msg.senderUsername === currentUser.name
                              ? "bg-[#4dbf38] text-white"
                              : "bg-gray-100"
                          }`}
                        >
                          <p className="text-sm">
                            {msg.senderUsername !== currentUser.name && (
                              <span className="font-semibold block mb-1">{msg.senderUsername}</span>
                            )}
                            {msg.plainText || msg.messages}
                          </p>
                          <span className="block text-xs opacity-75 mt-1">
                            {new Date(msg.timestamp).toLocaleString()}
                          </span>
                        </div>
                      </div>
                    ))}
                  </CardContent>
                  <div className="p-4 border-t border-gray-200 bg-white">
                    <div className="flex items-center space-x-2">
                      <Input
                        value={newMessage}
                        onChange={(e) => setNewMessage(e.target.value)}
                        placeholder="Type a message..."
                        className="flex-grow"
                        disabled={loading.sending}
                      />
                      <Button
                        onClick={sendMessage}
                        disabled={loading.sending || !newMessage.trim()}
                        className="bg-[#4dbf38] hover:bg-[#80d12a] text-white transition-colors duration-200"
                      >
                        {loading.sending ? (
                          <Loader className="h-4 w-4 animate-spin" />
                        ) : (
                          <Send className="h-4 w-4" />
                        )}
                        <span className="ml-2">{loading.sending ? "Sending..." : "Send"}</span>
                      </Button>
                    </div>
                  </div>
                </>
              ) : (
                <div className="flex items-center justify-center h-full text-gray-500">
                  Select a user to start chatting
                </div>
              )}
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default ChatApp;

