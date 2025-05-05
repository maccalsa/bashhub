# 📌 bashhub

> **A powerful TUI and CLI tool for managing, executing, and sharing your bash scripts effortlessly.**

`bashhub` provides an intuitive Terminal User Interface (TUI) and flexible CLI to store, categorize, run, and manage your bash scripts, complete with dynamic placeholders, syntax highlighting, and real-time search.

---

## 🎯 **Key Features**

* **TUI (Terminal User Interface)** for quick script selection, editing, and execution.
* **Dynamic placeholders** prompt for user input when executing scripts.
* **Automatic syntax highlighting** for enhanced readability (bash, YAML, JSON, etc.).
* **Category-based organization** to neatly group scripts.
* **Instant real-time filtering** to find scripts quickly.
* **Single-script export** with placeholder substitution.
* **Seamless CLI integration** for scripting and automation workflows.

---

## 🚀 **Quick Start**

### 🔧 **Installation**

```bash
git clone https://github.com/maccalsa/bashhub.git
cd bashhub
make install
```

Ensure your `$EDITOR` environment variable is set (default is `nano` if unset):

```bash
export EDITOR=nano  # or vim, emacs, code
```

### ▶️ **Run the TUI**

```bash
bashhub
```

### 📟 **Use the CLI**

Run a stored script directly from the command line:

```bash
bashhub run <script-name> --set placeholder=value
```

Export a script with placeholders substituted to your terminal:

```bash
bashhub export <script-name> --set placeholder=value
```

---

## 📖 **Using the TUI (Interactive Mode)**

* **Navigate** between panes with `Tab` / `Shift+Tab`.
* **Execute** a selected script by pressing `E`.
* **Create** a new script with `C`.
* **Edit** a selected script with `X`.
* **Delete** a script with `D`.
* **Search** scripts with `/`, press `Enter` to confirm, or `Esc` to cancel.
* **Exit** the app clearly using `Ctrl+Q`.

### 🗂️ **Organizing Your Scripts**

Scripts can be categorized for clear organization:

* Easily create new categories when saving or editing scripts.
* Categories and scripts within are automatically sorted alphabetically.

---

## 🎨 **Syntax Highlighting & Detection**

`bashhub` automatically highlights your scripts based on their content:

* Bash/Shell scripts (`.sh`)
* JSON files (`.json`)
* YAML files (`.yaml` or `.yml`)
* Python scripts (`.py`)
* And more...

---

## 🔖 **Using Placeholders**

Embed placeholders clearly within your scripts for dynamic inputs:

```bash
echo "Connecting to server {{server}} as user {{user}}"
ssh {{user}}@{{server}}
```

When executing, you'll clearly be prompted for inputs in the TUI or via CLI flags.

### Example (CLI):

```bash
bashhub run ssh-connect --set user=bob --set server=example.com
```

---

## 📤 **Exporting Scripts**

Export scripts with placeholders substituted directly to your terminal or redirect to files:

```bash
bashhub export backup-script --set dir=/home/user/data > backup.sh
chmod +x backup.sh
```

---

## 📂 **Bulk Importing**

Quickly import scripts from an existing folder:

```bash
bashhub import ./scripts-folder
```

This clearly imports all scripts at the root level of the specified folder (no nested folders).

---

## 🎖️ **Roadmap & Upcoming Features**

* Execution history and analytics
* Secret management integration
* Git-backed script versioning and collaboration
* Advanced placeholder types (validation, choices)
* User-defined customization and themes
* Enhanced CLI scripting and CI/CD integration

---

## 📝 **Contributing**

Contributions are welcome! Open issues, submit pull requests, or discuss ideas.

---

## 📜 **License**

MIT License © Your Name

---

## 🙌 **Thanks for using bashhub!**.

