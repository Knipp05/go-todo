import { useState } from "react";
import ToDoForm from "./ToDoForm";
import ToDo from "./ToDo";
import { Button } from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import { BASE_URL, Task, User } from "../App";
import "../App.css";

export default function Planer(props: any) {
  const [showForm, setShowForm] = useState(false);
  const [taskData, setTaskData] = useState<Task | null>(null);
  const [formType, setFormType] = useState("create");
  const [taskOrder, setTaskOrder] = useState(0);

  function handleCreateClick() {
    setTaskData(null);
    setFormType("create");
    setShowForm(true);
  }

  function handleEditClick(task: Task, order: number) {
    setTaskOrder(order);
    setTaskData(task);
    setFormType("edit");
    setShowForm(true);
  }

  function handleShareClick(task: Task) {
    setTaskData(task);
    setFormType("share");
    setShowForm(true);
  }
  const swapTasks = async (index: number) => {
    const token = sessionStorage.getItem("token");
    if (token && (index >= 0 || index < props.user.tasks.length - 1)) {
      {
        try {
          const res = await fetch(
            BASE_URL +
              `/tasks/${props.user.tasks[index].id}/${
                props.user.tasks[index + 1].id
              }`,
            {
              method: "PATCH",
              headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${token}`,
              },
            }
          );

          if (!res.ok) {
            const data = await res.json();
            throw new Error(data.error || "Unbekannter Fehler aufgetreten");
          }
        } catch (error: any) {
          console.log("Fehler bei Änderung der Reihenfolge:", error.message);
        }
      }

      var newTasks: Task[] = [...props.user.tasks];

      [newTasks[index], newTasks[index + 1]] = [
        newTasks[index + 1],
        newTasks[index],
      ];

      props.setUser((oldUser: User) => {
        return { ...oldUser, tasks: newTasks };
      });
    }
  };
  const todoElements = props.user.tasks.map((task: Task, idx: number) => (
    <ToDo
      key={idx}
      order={idx}
      data={task}
      user={props.user}
      setUser={props.setUser}
      handleEditClick={handleEditClick}
      handleShareClick={handleShareClick}
      swapTasks={swapTasks}
    />
  ));

  return (
    <div>
      <Button
        onClick={handleCreateClick}
        variant="contained"
        sx={{ marginTop: 2, marginLeft: "45%" }}
      >
        <AddIcon /> Aufgabe erstellen
      </Button>
      <ToDoForm
        open={showForm}
        setShowForm={setShowForm}
        setUser={props.setUser}
        user={props.user}
        type={formType}
        data={taskData}
        taskOrder={taskOrder}
      />
      {todoElements}
    </div>
  );
}
