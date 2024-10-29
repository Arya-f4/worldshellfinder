<?php 

?>
<!DOCTYPE html>
<html>
<head>
    <title>World Shell finder</title>
    <link rel="stylesheet" type="text/css" href="style.css">
</head>
<body>
<div class="container">
        <div class="header">
            <h2>World Shell Finder</h2>
        </div>
        <div class="content">
            <form action="" method="post">
                <div class="input-group">
                    <label>Directory</label>
                    <input type="text" name="dir">
                </div>
                <div class="input-group">
                    <button type="submit" class="btn" name="run_command">Run Command</button>
                </div>
            </form>
            <div class="output">
                <?php 
                    if (isset($_POST['run_command'])) {
                        $dir = escapeshellarg($_POST['dir']);
                        $command = "./worldfinder_linux_amd64 $dir";
                        $output = shell_exec($command);
                        if ($output === null) {
                            echo "Failed to execute command. Please check server permissions and file path.";
                        } else {
                            echo "<pre>" . htmlspecialchars($output) . "</pre>";
                        }
                    }
                ?>
            </div>
        </div>
    </div>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #f4f4f4;
        }
        .container {
            width: 50%;
            margin: 0 auto;
            padding: 20px;
            background-color: #fff;
            border-radius: 5px;
            box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
            margin-top: 50px;
        }
        .header {
            background-color: #333;
            color: #fff;
            padding: 10px;
            text-align: center;
            border-radius: 5px 5px 0 0;
        }
        .content {
            padding: 20px;
        }
        .input-group {
            margin: 10px 0;
        }
        .input-group label {
            display: block;
            font-weight: bold;
        }
        .input-group input {
            width: 100%;
            padding: 5px;
            border: 1px solid #ccc;
            border-radius: 5px;
        }
        .btn {
            padding: 5px 10px;
            background-color: #333;
            color: #fff;
            border: none;
            border-radius: 5px;
            cursor: pointer;
        }
        .output {
            margin-top: 20px;
            padding: 10px;
            background-color: #f9f9f9;
            border: 1px solid #ccc;
            border-radius: 5px;
            min-height: 200px;
            overflow: auto;
        }

        </style>
</body>
</html>
