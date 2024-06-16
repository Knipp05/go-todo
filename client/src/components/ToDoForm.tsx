import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  TextField,
  Grid,
  IconButton,
} from "@mui/material";
import { useEffect, useState } from "react";
import { BASE_URL, Task, User } from "../App";
import CategoryMenu from "./CategoryMenu";
import ClearIcon from "@mui/icons-material/Clear";

export default function ToDoForm(props: any) {
  const [errorMessage, setErrorMessage] = useState("");
  const [taskContent, setTaskContent] = useState({
    title: "",
    desc: "",
    isDone: false,
    owner: props.user.name,
    shared: [""],
    order: props.user.tasks.length + 1,
    category: {
      id: 1,
      cat_name: "default",
      color_header: "#00a4ba",
      color_body: "#00ceea",
    },
  });
  const [targetName, setTargetName] = useState("");

  useEffect(() => {
    if (props.type !== "create" && props.data) {
      setTaskContent({
        title: props.data.title,
        desc: props.data.desc,
        isDone: props.data.isDone,
        owner: props.data.owner,
        shared: props.data.shared,
        order: props.data.order,
        category: props.data.category,
      });
    } else if (props.type === "create") {
      setTaskContent({
        title: "",
        desc: "",
        isDone: false,
        owner: props.user.name,
        shared: [""],
        order: props.user.tasks.length + 1,
        category: {
          id: 1,
          cat_name: "default",
          color_header: "#00a4ba",
          color_body: "#00ceea",
        },
      });
    }
  }, [props.open]);

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
    setTaskContent({
      title: "",
      desc: "",
      isDone: false,
      owner: props.user.name,
      shared: [""],
      order: props.user.tasks.length + 1,
      category: {
        id: 1,
        cat_name: "default",
        color_header: "#00a4ba",
        color_body: "#00ceea",
      },
    });
    if (props.type === "share") {
      setTargetName("");
      setErrorMessage("");
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
    if (props.type !== "share") {
      setTaskContent((oldContent) => {
        return { ...oldContent, [event.target.name]: event.target.value };
      });
    } else {
      setTargetName(event.target.value);
    }
  }

  const changeTask = async () => {
    const token = sessionStorage.getItem("token");
    const target =
      props.type === "create"
        ? `/${props.user.name}/tasks`
        : `/${props.user.name}/tasks/${props.data.id}`;
    const method = props.type === "create" ? "POST" : "PATCH";
    if (token && taskContent.title.trim() !== "") {
      try {
        const res = await fetch(BASE_URL + target, {
          method: method,
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
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
                owner: taskContent.owner,
                shared: [""],
                order: taskContent.order,
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
        handleClose();
      } catch (error: any) {
        throw new Error("Fehler beim Erstellen/Ändern der Aufgabe aufgetreten");
      }
    }
  };
  const shareTask = async () => {
    const token = sessionStorage.getItem("token");
    if (token && targetName.trim() !== "") {
      try {
        const res = await fetch(
          BASE_URL + `/${props.user.name}/tasks/${props.data.id}/${targetName}`,
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              Authorization: `Bearer ${token}`,
            },
            body: JSON.stringify({
              ...taskContent,
              shared:
                taskContent.shared[0] === ""
                  ? [targetName]
                  : [...taskContent.shared, targetName],
            }),
          }
        );

        if (!res.ok) {
          const data = await res.json();
          setErrorMessage(data.error);
          throw new Error(data.error || "Unbekannter Fehler aufgetreten");
        }
        props.setUser((oldUser: User) => {
          const updatedTasks = oldUser.tasks.map((task: Task) => {
            if (task.id === props.data.id) {
              task.shared[0] === ""
                ? (task.shared[0] = targetName)
                : task.shared.push(targetName);
              return task;
            } else {
              return task;
            }
          });
          return { ...oldUser, tasks: updatedTasks };
        });
        handleClose();
      } catch (error: any) {
        throw new Error("Fehler beim Freigeben der Aufgabe aufgetreten");
      }
    }
  };

  const handleRemoveShare = async (target: string) => {
    const token = sessionStorage.getItem("token");
    if (token) {
      try {
        const res = await fetch(
          BASE_URL + `/${props.user.name}/tasks/${props.data.id}/${target}`,
          {
            method: "DELETE",
            headers: {
              "Content-Type": "application/json",
              Authorization: `Bearer ${token}`,
            },
          }
        );

        if (!res.ok) {
          const data = await res.json();
          setErrorMessage(data.error);
          throw new Error(data.error || "Unbekannter Fehler aufgetreten");
        }
        props.setUser((oldUser: User) => {
          const updatedTasks = oldUser.tasks.map((task: Task) => {
            if (task.id === props.data.id) {
              const updatedShares = task.shared.filter(
                (share) => share !== target
              );
              task.shared = updatedShares;
              return task;
            } else {
              return task;
            }
          });
          return { ...oldUser, tasks: updatedTasks };
        });
        handleClose();
      } catch (error: any) {
        throw new Error("Fehler beim Freigeben der Aufgabe aufgetreten");
      }
    }
  };

  const sharedUsers =
    props.data !== null
      ? props.data.shared.map((share: string, idx: number) => (
          <div key={idx} style={{ marginLeft: "4%" }}>
            {share}{" "}
            <IconButton size="small" onClick={() => handleRemoveShare(share)}>
              <ClearIcon />
            </IconButton>
          </div>
        ))
      : null;

  return (
    <Dialog open={props.open} onClose={handleClose}>
      <DialogContent>
        {props.type !== "share" && (
          <DialogTitle>
            {props.type === "create"
              ? "Neue Aufgabe anlegen"
              : "Aufgabe bearbeiten"}
          </DialogTitle>
        )}
        {props.type === "share" && (
          <DialogTitle>Aufgabe für anderen Benutzer freigeben</DialogTitle>
        )}
        <DialogContentText>
          {props.type !== "share"
            ? "Bitte gib einen Titel und optional eine Beschreibung für die Aufgabe an"
            : "Bitte gib den Benutzernamen ein, mit dem du diese Aufgabe teilen möchtest"}
        </DialogContentText>
        <Grid container direction="column" spacing={2}>
          <Grid item>
            <TextField
              required
              id="title"
              name="title"
              label={props.type !== "share" ? "Titel" : "Benutzer"}
              type="text"
              variant="standard"
              inputProps={{ maxLength: 40 }}
              value={props.type !== "share" ? taskContent.title : targetName}
              onChange={handleInput}
              fullWidth
            />
          </Grid>
          {props.type === "share" && (
            <DialogContentText
              sx={{ color: "red", marginLeft: "2.5%", marginTop: "1%" }}
            >
              {errorMessage}
            </DialogContentText>
          )}
          {props.type === "share" && props.data.shared[0] !== "" && (
            <DialogTitle>Aufgabe bereits freigegeben für:</DialogTitle>
          )}
          {props.type === "share" && props.data.shared[0] !== "" && sharedUsers}
          {props.type !== "share" && (
            <Grid item>
              <TextField
                id="desc"
                name="desc"
                label="Beschreibung"
                type="text"
                variant="standard"
                inputProps={{ maxLength: 128 }}
                value={taskContent.desc}
                onChange={handleInput}
                fullWidth
              />
            </Grid>
          )}
          {props.type !== "share" && (
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
          )}
        </Grid>
        {props.type !== "share" && (
          <CategoryMenu
            changeCategory={changeCategory}
            open={open}
            handleClose={handleDropDownClose}
            anchorEl={anchorEl}
            categories={props.categories}
            user={props.user}
            setUser={props.setUser}
            setTaskContent={setTaskContent}
          />
        )}
        <DialogActions>
          <Button onClick={handleClose}>Abbrechen</Button>
          {props.type !== "share" && (
            <Button onClick={changeTask}>
              {props.type === "create"
                ? "Aufgabe erstellen"
                : "Änderungen speichern"}
            </Button>
          )}
          {props.type === "share" && (
            <Button onClick={shareTask}>Aufgabe freigeben</Button>
          )}
        </DialogActions>
      </DialogContent>
    </Dialog>
  );
}
