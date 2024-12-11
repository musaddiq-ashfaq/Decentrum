import { useCallback, useEffect, useState } from "react";
import { toast, Toaster } from "sonner";
import { Button } from "../Components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../Components/ui/card";
import { Input } from "../Components/ui/input";

const API_BASE_URL = "http://localhost:8081";

const GroupChatApp = () => {
  // State management
  const [currentUser, setCurrentUser] = useState(null);
  const [groups, setGroups] = useState([]);
  const [selectedGroup, setSelectedGroup] = useState(null);
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState("");
  const [loading, setLoading] = useState({
    groups: false,
    messages: false,
    sending: false,
  });

  // Retrieve user from local storage
  const retrieveUserFromStorage = useCallback(() => {
    try {
      const userDataString = localStorage.getItem("user");
      if (userDataString) {
        const userData = JSON.parse(userDataString);
        setCurrentUser(userData);
        return userData;
      }
      return null;
    } catch (error) {
      console.error("Error parsing user data:", error);
      localStorage.removeItem("user");
      return null;
    }
  }, []);

  // Fetch groups
  const fetchGroups = useCallback(async () => {
    const user = currentUser || retrieveUserFromStorage();
    if (!user) {
      toast.error("Please log in to view groups");
      return;
    }

    setLoading((prev) => ({ ...prev, groups: true }));
    try {
      const response = await fetch(`${API_BASE_URL}/usergroups`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ user_name: user.name }),
      });

      if (!response.ok) {
        throw new Error("Failed to fetch groups");
      }

      const data = await response.json();
      const fetchedGroups = data.groups || [];
      console.log("Fetched Groups:", fetchedGroups);
      setGroups(fetchedGroups);

      // Auto-select first group if available
      if (fetchedGroups.length > 0 && !selectedGroup) {
        setSelectedGroup(fetchedGroups[0]);
      }
    } catch (error) {
      console.error("Groups fetch error:", error);
      toast.error("Could not retrieve groups");
    } finally {
      setLoading((prev) => ({ ...prev, groups: false }));
    }
  }, [currentUser, retrieveUserFromStorage, selectedGroup]);

  // Fetch messages from localStorage or the server
  const fetchMessages = useCallback(async () => {
    if (!currentUser || !selectedGroup) return;

    // Check if messages are stored in localStorage
    const storedMessages = JSON.parse(localStorage.getItem(`messages_${selectedGroup.id}`));
    if (storedMessages) {
      setMessages(storedMessages);
      return;
    }

    setLoading((prev) => ({ ...prev, messages: true }));

    try {
      const response = await fetch(`${API_BASE_URL}/groupchat`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          operation: "get",
          groupID: selectedGroup.id,
          participants: selectedGroup.members,
          senderUsername: currentUser.name,
        }),
      });

      if (!response.ok) {
        throw new Error("Failed to fetch messages");
      }

      // Retrieve all messages and process them
      const messageData = await response.json();
      const members = selectedGroup.members || []; // Assuming `members` is part of the group data
      const allMessages = [];

      for (const member of members) {
        const processedMessages = messageData.map((msg) => ({
          senderUsername: member,
          plainText: msg,
          timestamp: msg.timestamp || new Date().toISOString(),
        }));

        allMessages.push(...processedMessages);
      }

      // Sort messages by timestamp to display in the correct order
      allMessages.sort(
        (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
      );

      setMessages(allMessages);

      // Store messages in localStorage for future use
      localStorage.setItem(`messages_${selectedGroup.id}`, JSON.stringify(allMessages));
    } catch (error) {
      console.error("Messages fetch error:", error);
      toast.error("Could not retrieve messages");
    } finally {
      setLoading((prev) => ({ ...prev, messages: false }));
    }
  }, [currentUser, selectedGroup]);

  // Send message
  const sendMessage = async () => {
    if (!newMessage.trim() || !currentUser || !selectedGroup) return;

    setLoading((prev) => ({ ...prev, sending: true }));
    try {
      const response = await fetch(`${API_BASE_URL}/groupchat`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          operation: "send",
          participants: selectedGroup.members,
          groupID: selectedGroup.id,
          username: currentUser.name,
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

      setMessages((prev) => [...prev, newMessageObj]);
      setNewMessage("");

      // Store the new message in localStorage
      const updatedMessages = [...messages, newMessageObj];
      localStorage.setItem(`messages_${selectedGroup.id}`, JSON.stringify(updatedMessages));

      // Refresh messages to ensure sync with backend
      await fetchMessages();
    } catch (error) {
      console.error("Send message error:", error);
      toast.error("Could not send message");
    } finally {
      setLoading((prev) => ({ ...prev, sending: false }));
    }
  };

  // Effects for user data, groups, and message fetching
  useEffect(() => {
    retrieveUserFromStorage();
  }, [retrieveUserFromStorage]);

  useEffect(() => {
    if (currentUser) {
      fetchGroups();
    }
  }, [currentUser, fetchGroups]);

  useEffect(() => {
    let intervalId = null;
    if (selectedGroup) {
      fetchMessages();
      intervalId = setInterval(fetchMessages, 5000);
    }
    return () => {
      if (intervalId) clearInterval(intervalId);
    };
  }, [selectedGroup, fetchMessages]);

  return (
    <div className="flex h-screen">
      <Toaster richColors />

      {/* Sidebar: Groups List */}
      <div className="w-1/4 bg-gray-100 p-4 overflow-y-auto">
        <h2 className="text-xl font-bold mb-4">Your Groups</h2>
        {loading.groups ? (
          <p className="text-gray-500">Loading groups...</p>
        ) : groups.length === 0 ? (
          <p className="text-gray-500">No groups found</p>
        ) : (
          groups.map((group) => (
            <div
              key={group.groupID}
              className={`p-2 mb-2 cursor-pointer rounded ${
                selectedGroup?.id === group.groupID
                  ? "bg-blue-500 text-white"
                  : "hover:bg-gray-200"
              }`}
              onClick={() => setSelectedGroup(group)}
            >
              {group.groupname}
            </div>
          ))
        )}
      </div>

      {/* Chat Window */}
      <div className="w-3/4 flex flex-col">
        {selectedGroup ? (
          <Card className="h-full flex flex-col">
            <CardHeader>
              <CardTitle>
                {selectedGroup.groupname}
                {loading.messages && (
                  <span className="ml-2 text-sm text-gray-500">
                    Loading messages...
                  </span>
                )}
              </CardTitle>
            </CardHeader>

            <CardContent className="flex-grow overflow-y-auto flex flex-col">
              {messages.map((msg, index) => (
                <div
                  key={index}
                  className={`mb-2 p-2 rounded max-w-xs ${
                    msg.senderUsername === currentUser?.name
                      ? "bg-blue-100 self-end text-right ml-auto"
                      : "bg-gray-100 self-start text-left mr-auto"
                  }`}
                >
                  <p className="text-sm text-gray-700">
                    {msg.senderUsername !== currentUser?.name && (
                      <span className="font-semibold mr-1">
                        {msg.senderUsername}:
                      </span>
                    )}
                    {msg.plainText}
                  </p>
                  <span className="text-xs text-gray-500">{new Date(msg.timestamp).toLocaleTimeString()}</span>
                </div>
              ))}
            </CardContent>

            {/* Message Input */}
            <div className="flex p-4">
              <Input
                value={newMessage}
                onChange={(e) => setNewMessage(e.target.value)}
                placeholder="Type a message..."
                className="flex-grow mr-2"
              />
              <Button
                disabled={loading.sending || !newMessage.trim()}
                onClick={sendMessage}
              >
                {loading.sending ? "Sending..." : "Send"}
              </Button>
            </div>
          </Card>
        ) : (
          <div className="flex-grow flex items-center justify-center">
            <p className="text-gray-500">Select a group to start chatting</p>
          </div>
        )}
      </div>
    </div>
  );
};

export default GroupChatApp;
