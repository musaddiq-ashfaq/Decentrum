import { ArrowUpRight } from "lucide-react";
import { Badge } from "./badge";
import { Button } from "./button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "./card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "./table";
import { Link } from "react-router-dom/cjs/react-router-dom.min";
import axios from "axios";
import { useEffect, useState } from "react";

export function CardWithForm() {
  const [Chats, setChats] = useState([]);

  const fetchChats = async () => {
    const { data } = await axios.get("/api/chats");
    setChats(data);
  };

  useEffect(() => {
    fetchChats();
  }, []);
  return (
    <Card>
      <CardHeader>
        <div>
          <CardTitle>Chats</CardTitle>
          <CardDescription></CardDescription>
        </div>
      </CardHeader>
      <CardContent>
      {Chats.map((chat) => {
              return (
                <div key={chat._id} className="p-2 sm:p-3 md:p-5">
                  <Card>
                    <CardHeader>
                        <div>
                            <CardTitle>{chat.chatName}</CardTitle>
                        </div>
                    </CardHeader>
                    <CardContent>
                        <p>{chat._id}</p>
                    </CardContent>
                  </Card>
                </div>
              );
            })}
      </CardContent>
    </Card>
  );
}
