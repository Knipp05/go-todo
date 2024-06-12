import { IconButton } from "@mui/material";
import DeleteIcon from "@mui/icons-material/Delete";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import EditIcon from "@mui/icons-material/Edit";
import ShareIcon from "@mui/icons-material/Share";
import "../App.css";
import { BASE_URL, Task, User } from "../App";

export default function ToDo(props: any) {
  const handleDelete = async (taskId: number) => {
    const token = sessionStorage.getItem("token");
    if (token) {
      try {
        const res = await fetch(
          BASE_URL + `/${props.user.name}/tasks/${taskId}`,
          {
            method: "DELETE",
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
          BASE_URL + `/${props.user.name}/tasks/${taskid}/${props.data.isDone}`,
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
        {props.user.name === props.data.owner && (
          <IconButton
            size="small"
            onClick={() => props.handleShareClick(props.data)}
          >
            <ShareIcon />
          </IconButton>
        )}
        {props.user.name === props.data.owner && (
          <IconButton
            size="small"
            onClick={() => props.handleEditClick(props.data)}
          >
            <EditIcon />
          </IconButton>
        )}
      </div>
      <div className="todo--desc">
        <p>{props.data.desc}</p>
        {props.user.name === props.data.owner && (
          <IconButton size="small" onClick={() => handleDelete(props.data.id)}>
            <DeleteIcon />
          </IconButton>
        )}
        <IconButton size="small" onClick={() => handleCheck(props.data.id)}>
          <CheckCircleIcon />
        </IconButton>
      </div>
    </div>
  );
}
