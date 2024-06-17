import {
  Menu,
  Divider,
  MenuItem,
  Dialog,
  DialogTitle,
  TextField,
  DialogActions,
  Button,
  DialogContentText,
  DialogContent,
  Grid,
  IconButton,
} from "@mui/material";
import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import { BASE_URL, Category, User } from "../App";
import { useState } from "react";
import ColorPicker from "./ColorPicker";

export default function CategoryMenu(props: any) {
  const [dialogType, setDialogType] = useState("create");
  const [category, setCategory] = useState<Category>({
    id: 1,
    cat_name: "",
    color_header: "#00a4ba",
    color_body: "#00ceea",
  });
  const [showDialog, setShowDialog] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");

  function handleInput(event: any) {
    setCategory((oldCategory) => {
      return { ...oldCategory, [event.target.name]: event.target.value };
    });
  }
  function handleClose() {
    setDialogType("create");
    setShowDialog(false);
    setErrorMessage("");
    setCategory({
      id: 1,
      cat_name: "",
      color_header: "#00a4ba",
      color_body: "#00ceea",
    });
  }
  const submitInput = async () => {
    const token = sessionStorage.getItem("token");
    if (
      token &&
      category.cat_name !== "default" &&
      category.cat_name.trim() !== ""
    ) {
      const path =
        dialogType === "create" ? `/categories` : `/categories/${category.id}`;
      const method = dialogType === "create" ? "POST" : "PATCH";
      try {
        const res = await fetch(BASE_URL + path, {
          method: method,
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify(category),
        });

        const data = await res.json();

        if (!res.ok) {
          throw new Error(data.error || "Unbekannter Fehler aufgetreten");
        }
        if (dialogType === "create") {
          props.setUser((oldUser: User) => {
            var newCategories = [
              ...oldUser.categories,
              {
                id: data.id,
                cat_name: category.cat_name,
                color_header: category.color_header,
                color_body: category.color_body,
              },
            ];
            return { ...oldUser, categories: newCategories };
          });
          props.setTaskContent((oldContent: any) => {
            return {
              ...oldContent,
              category: {
                id: data.id,
                cat_name: category.cat_name,
                color_header: category.color_header,
                color_body: category.color_body,
              },
            };
          });
        } else {
          props.setUser((oldUser: User) => {
            const updatedCategories = oldUser.categories.map(
              (cat: Category) => {
                if (cat.id === category.id) {
                  cat.cat_name = category.cat_name;
                  cat.color_header = category.color_header;
                  cat.color_body = category.color_body;
                  return cat;
                } else {
                  return cat;
                }
              }
            );
            return { ...oldUser, categories: updatedCategories };
          });
        }

        props.changeCategory(
          data.id,
          category.cat_name,
          category.color_header,
          category.color_body
        );
        handleClose();
        props.handleClose();
      } catch (error: any) {
        setErrorMessage(error);
        throw new Error(
          "Fehler beim Erstellen/Ändern der Kategorie aufgetreten"
        );
      }
    } else {
      setErrorMessage("Kategoriename ungültig");
    }
  };

  function editCategory(category: Category) {
    setCategory(category);
    setDialogType("edit");
    setShowDialog(true);
  }

  const deleteCategory = async (id: number) => {
    const token = sessionStorage.getItem("token");
    if (token) {
      try {
        const res = await fetch(BASE_URL + `/categories/${id}/delete`, {
          method: "PATCH",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
          },
        });

        const data = await res.json();
        if (!res.ok) {
          throw new Error(data.error || "Unbekannter Fehler aufgetreten");
        }
        props.setUser((oldUser: User) => {
          const updatedCategories = oldUser.categories.filter(
            (category: Category) => category.id !== id
          );
          return {
            ...oldUser,
            tasks: data.tasks,
            categories: updatedCategories,
          };
        });
        props.changeCategory(1, "default", "#00a4ba", "#00ceea");
      } catch (error: any) {
        console.log("Fehler beim Löschen der Aufgabe:", error.message);
      }
    }
  };

  const categoryElements = props.categories.map(
    (category: Category, idx: number) => (
      <MenuItem
        key={idx}
        sx={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
        onClick={() => {
          props.changeCategory(
            category.id,
            category.cat_name,
            category.color_header,
            category.color_body
          );
          props.handleClose();
        }}
      >
        {category.cat_name !== "default"
          ? category.cat_name
          : "nicht kategorisiert"}
        <div>
          {category.cat_name !== "default" && (
            <IconButton size="small" onClick={() => editCategory(category)}>
              <EditIcon fontSize="small" />
            </IconButton>
          )}
          {category.cat_name !== "default" && (
            <IconButton
              size="small"
              onClick={() => deleteCategory(category.id)}
            >
              <DeleteIcon fontSize="small" />
            </IconButton>
          )}
        </div>
      </MenuItem>
    )
  );

  return (
    <>
      <Menu
        id="basic-menu"
        anchorEl={props.anchorEl}
        open={props.open}
        onClose={props.handleClose}
        MenuListProps={{
          "aria-labelledby": "basic-button",
        }}
      >
        {categoryElements}
        <Divider sx={{ my: 0.5 }} />
        <MenuItem
          onClick={() => {
            setShowDialog(true);
            props.handleClose();
          }}
        >
          Kategorie hinzufügen
        </MenuItem>
      </Menu>
      <Dialog open={showDialog} onClose={handleClose}>
        <DialogTitle>
          {dialogType === "create"
            ? "Neue Kategorie anlegen"
            : "Kategorie bearbeiten"}
        </DialogTitle>
        <DialogContent>
          <DialogContentText>
            Bitte Kategorienamen und optional eine Farbe angeben
          </DialogContentText>
          <TextField
            required
            id="cat_name"
            name="cat_name"
            label="Name"
            type="text"
            variant="standard"
            inputProps={{ maxLength: 14 }}
            value={category.cat_name}
            onChange={handleInput}
            fullWidth
            margin="normal"
          />
          {errorMessage !== "" && (
            <DialogContentText sx={{ color: "red" }}>
              {errorMessage}
            </DialogContentText>
          )}
          <Grid container spacing={2}>
            <Grid item xs={6}>
              <ColorPicker
                setCategory={setCategory}
                categoryColor={category.color_header}
                dialogType={dialogType}
                type="color_header"
              />
            </Grid>
            <Grid item xs={6}>
              <ColorPicker
                setCategory={setCategory}
                categoryColor={category.color_body}
                dialogType={dialogType}
                type="color_body"
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose}>Abbrechen</Button>
          <Button onClick={submitInput}>
            {dialogType === "create"
              ? "Kategorie anlegen"
              : "Änderungen speichern"}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}
