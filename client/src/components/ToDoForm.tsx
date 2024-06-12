import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  TextField,
  Grid,
} from "@mui/material";
import { useEffect, useState } from "react";
import { BASE_URL, Task, User } from "../App";
import CategoryMenu from "./CategoryMenu";

export default function ToDoForm(props: any) {
  const [taskContent, setTaskContent] = useState({
    title: "",
    desc: "",
    category: {
      id: 1,
      cat_name: "default",
      color_header: "#00a4ba",
      color_body: "#00ceea",
    },
  });

  useEffect(() => {
    if (props.type === "edit" && props.data) {
      setTaskContent({
        title: props.data.title,
        desc: props.data.desc,
        category: props.data.category,
      });
    } else if (props.type === "create") {
      setTaskContent({
        title: "",
        desc: "",
        category: {
          id: 1,
          cat_name: "default",
          color_header: "#00a4ba",
          color_body: "#00ceea",
        },
      });
    }
  }, [props.data, props.type]);

  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  const handleClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleDropDownClose = () => {
    setAnchorEl(null);
  };

  function handleClose() {
    props.setShowForm(false);
    if (props.type === "edit" && props.data) {
      setTaskContent({
        title: props.data.title,
        desc: props.data.desc,
        category: props.data.category,
      });
    } else if (props.type === "create") {
      setTaskContent({
        title: "",
        desc: "",
        category: {
          id: 1,
          cat_name: "default",
          color_header: "#00a4ba",
          color_body: "#00ceea",
        },
      });
    }
  }

  function changeCategory(
    id: number,
    cat_name: string,
    color_header: string,
    color_body: string
  ) {
    setTaskContent((oldContent) => {
      return {
        ...oldContent,
        category: {
          id: id,
          cat_name: cat_name,
          color_header: color_header,
          color_body: color_body,
        },
      };
    });
  }

  function handleInput(event: any) {
    setTaskContent((oldContent) => {
      return { ...oldContent, [event.target.name]: event.target.value };
    });
  }

  const submitInput = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const token = sessionStorage.getItem("token");
    const target =
      props.type === "create"
        ? `/${props.user.id}/tasks`
        : `/${props.user.name}/tasks/${props.data.id}`;
    const method = props.type === "create" ? "POST" : "PATCH";
    if (token && taskContent.title !== "") {
      try {
        const res = await fetch(BASE_URL + target, {
          method: method,
          headers: {
            "Content-Type": "application/json",
            Authorization: token,
          },
          body: JSON.stringify(taskContent),
        });

        const data = await res.json();

        if (!res.ok) {
          throw new Error(data.error || "Unbekannter Fehler aufgetreten");
        }
        if (props.type === "create") {
          props.setUser((oldUser: User) => {
            var newTasks = [
              ...oldUser.tasks,
              {
                id: data.id,
                title: taskContent.title,
                desc: taskContent.desc,
                isDone: false,
                category: taskContent.category,
              },
            ];
            return { ...oldUser, tasks: newTasks };
          });
        } else {
          props.setUser((oldUser: User) => {
            const updatedTasks = oldUser.tasks.map((task: Task) => {
              if (task.id === props.data.id) {
                task.title = taskContent.title;
                task.desc = taskContent.desc;
                task.category = taskContent.category;
                return task;
              } else {
                return task;
              }
            });
            return { ...oldUser, tasks: updatedTasks };
          });
        }
        props.setShowForm(false);
      } catch (error: any) {
        throw new Error("Fehler beim Erstellen/Ändern der Aufgabe aufgetreten");
      }
    }
  };

  return (
    <Dialog
      open={props.open}
      onClose={handleClose}
      PaperProps={{
        component: "form",
        onSubmit: (event: React.FormEvent<HTMLFormElement>) =>
          submitInput(event),
      }}
    >
      <DialogContent>
        <DialogTitle>
          {props.type === "create"
            ? "Neue Aufgabe anlegen"
            : "Aufgabe bearbeiten"}
        </DialogTitle>
        <DialogContentText>
          Bitte gib einen Titel und optional eine Beschreibung für die Aufgabe
          an
        </DialogContentText>
        <Grid container direction="column" spacing={2}>
          <Grid item>
            <TextField
              required
              id="title"
              name="title"
              label="Titel"
              type="text"
              variant="standard"
              value={taskContent.title}
              onChange={handleInput}
              fullWidth
            />
          </Grid>
          <Grid item>
            <TextField
              id="desc"
              name="desc"
              label="Beschreibung"
              type="text"
              variant="standard"
              value={taskContent.desc}
              onChange={handleInput}
              fullWidth
            />
          </Grid>
          <Grid item>
            <Button
              id="basic-button"
              aria-controls={open ? "basic-menu" : undefined}
              aria-haspopup="true"
              aria-expanded={open ? "true" : undefined}
              onClick={handleClick}
            >
              {taskContent.category.cat_name !== "default"
                ? taskContent.category.cat_name
                : "nicht kategorisiert"}
            </Button>
          </Grid>
        </Grid>
        <CategoryMenu
          changeCategory={changeCategory}
          open={open}
          handleClose={handleDropDownClose}
          anchorEl={anchorEl}
          categories={props.categories}
          user={props.user}
          setUser={props.setUser}
        />
        <DialogActions>
          <Button onClick={handleClose}>Abbrechen</Button>
          <Button type="submit">
            {props.type === "create"
              ? "Aufgabe erstellen"
              : "Änderungen speichern"}
          </Button>
        </DialogActions>
      </DialogContent>
    </Dialog>
  );
}
