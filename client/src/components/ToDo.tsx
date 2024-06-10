import { Button } from "@mui/material";
import DeleteIcon from "@mui/icons-material/Delete";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import EditIcon from "@mui/icons-material/Edit";
import "../App.css";
import { BASE_URL, Task, User } from "../App";
import { useState } from "react";
import ToDoForm from "./ToDoForm";
export default function ToDo(props: any) {
  const [showForm, setShowForm] = useState(false);
  const handleDelete = async (taskId: number) => {
    const token = sessionStorage.getItem("token");
    if (token) {
      try {
        const res = await fetch(BASE_URL + `/tasks/${taskId}`, {
          method: "DELETE",
          headers: {
            "Content-Type": "application/json",
            Authorization: token,
          },
        });

        if (!res.ok) {
          const data = await res.json();
          throw new Error(data.error || "Unbekannter Fehler aufgetreten");
        }
        props.setUser((oldUser: User) => {
          const updatedTasks = oldUser.tasks.filter(
            (task: Task) => task.id !== taskId
          );
          return { ...oldUser, tasks: updatedTasks };
        });
      } catch (error: any) {
        console.log("Fehler beim Löschen der Aufgabe:", error.message);
      }
    }
  };
  const handleCheck = async (taskid: number) => {
    const token = sessionStorage.getItem("token");
    if (token) {
      try {
        const res = await fetch(
          BASE_URL + `/tasks/${taskid}/isdone/${props.data.isDone}`,
          {
            method: "PATCH",
            headers: {
              "Content-Type": "application/json",
              Authorization: token,
            },
          }
        );

        if (!res.ok) {
          const data = await res.json();
          throw new Error(data.error || "Unbekannter Fehler aufgetreten");
        }
        props.setUser((oldUser: User) => {
          const updatedTasks = oldUser.tasks.map((task: Task) => {
            if (task.id === props.data.id) {
              task.isDone = !task.isDone;
              return task;
            } else {
              return task;
            }
          });
          return { ...oldUser, tasks: updatedTasks };
        });
      } catch (error: any) {
        console.log("Fehler bei Statusänderung:", error.message);
      }
    }
  };
  return (
    <div
      className={props.data.isDone ? "todo todo--done" : "todo"}
      style={{ backgroundColor: props.data.category.color_body }}
    >
      <div
        className="todo--title"
        style={{ backgroundColor: props.data.category.color_header }}
      >
        <h2>{props.data.title}</h2>
        <h3>
          {props.data.category.cat_name === "default"
            ? "nicht kategorisiert"
            : props.data.category.cat_name}
        </h3>
        <Button onClick={() => setShowForm(true)}>
          <EditIcon />
        </Button>
      </div>
      <div className="todo--desc">
        <p>{props.data.desc}</p>
        <Button onClick={() => handleDelete(props.data.id)}>
          <DeleteIcon />
        </Button>
        <Button onClick={() => handleCheck(props.data.id)}>
          <CheckCircleIcon />
        </Button>
      </div>
      <ToDoForm
        open={showForm}
        setShowForm={setShowForm}
        setUser={props.setUser}
        data={props.data}
        user={props.user}
        categories={props.categories}
        type="edit"
      />
    </div>
  );
}
