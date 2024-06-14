import { useEffect, useState } from "react";
import "./App.css";
import NavBar from "./components/NavBar";
import Planer from "./components/Planer";

export type Task = {
  id: number;
  title: string;
  desc: string;
  isDone: boolean;
  category: Category;
  owner: string;
  shared: string[];
};
export type User = {
  name: string;
  tasks: Task[];
  categories: Category[];
};
export type Category = {
  id: number;
  cat_name: string;
  color_header: string;
  color_body: string;
};
export const BASE_URL = "http://localhost:5000/api";
function App() {
  const [user, setUser] = useState<User>({
    name: "",
    tasks: [],
    categories: [],
  });
  const [socket, setSocket] = useState<WebSocket | null>(null);

  useEffect(() => {
    if (!user.name) {
      return;
    }
    const token = sessionStorage.getItem("token");
    if (!token) return;
    const ws = new WebSocket(`ws://localhost:5000/ws?token=${token}`);
    setSocket(ws);

    ws.onmessage = (event) => {
      const updatedTask = JSON.parse(event.data);
      if (updatedTask.title) {
        setUser((oldUser) => {
          const index = oldUser.tasks.findIndex(
            (task) => task.id === updatedTask.id
          );
          if (index !== -1) {
            return {
              ...oldUser,
              tasks: oldUser.tasks.map((task) =>
                task.id === updatedTask.id ? updatedTask : task
              ),
            };
          } else {
            return {
              ...oldUser,
              tasks: [...oldUser.tasks, updatedTask],
            };
          }
        });
      } else {
        setUser((oldUser) => {
          return {
            ...oldUser,
            tasks: oldUser.tasks.filter((task) => task.id !== updatedTask),
          };
        });
      }
    };

    ws.onerror = (error) => {
      console.error("Websocket Error: ", error);
    };

    ws.onclose = () => {
      console.log("Websocket closed");
    };

    return () => {
      ws.close();
    };
  }, [user.name]);
  return (
    <div>
      <NavBar user={user} setUser={setUser} />
      {user.name !== "" && <Planer user={user} setUser={setUser} />}
      {user.name === "" && (
        <h1>Melde dich an, um deine Aufgaben zu sehen und zu verwalten</h1>
      )}
    </div>
  );
}

export default App;
