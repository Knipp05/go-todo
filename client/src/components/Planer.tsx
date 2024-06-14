import { useState } from "react";
import ToDoForm from "./ToDoForm";
import ToDo from "./ToDo";
import { Button } from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import { Task } from "../App";
import "../App.css";

export default function Planer(props: any) {
  const [showForm, setShowForm] = useState(false);
  const [taskData, setTaskData] = useState<Task | null>(null);
  const [formType, setFormType] = useState("create");

  function handleCreateClick() {
    setTaskData(null);
    setFormType("create");
    setShowForm(true);
  }

  function handleEditClick(task: Task) {
    setTaskData(task);
    setFormType("edit");
    setShowForm(true);
  }

  function handleShareClick(task: Task) {
    setTaskData(task);
    setFormType("share");
    setShowForm(true);
  }
  const todoElements = props.user.tasks.map((task: Task, idx: number) => (
    <ToDo
      key={idx}
      data={task}
      user={props.user}
      setUser={props.setUser}
      categories={props.user.categories}
      handleEditClick={handleEditClick}
      handleShareClick={handleShareClick}
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
        categories={props.user.categories}
        type={formType}
        data={taskData}
      />
      {todoElements}
    </div>
  );
}
