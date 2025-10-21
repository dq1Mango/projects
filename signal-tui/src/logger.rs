use std::fs::File;
use std::fs::OpenOptions;
use std::io::Write;

pub struct Logger {
  file: File,
}

impl Logger {
  pub fn init(file_name: &str) -> Self {
    let mut file = OpenOptions::new()
      .write(true)
      .truncate(true)
      .create(true)
      .open(file_name)
      .expect("am i goated?");

    let result = writeln!(file, "=== START OF LOG === ");

    result.expect("am i goated?");

    file = OpenOptions::new().append(true).create(true).open(file_name).expect("kaboom");

    let logger = Logger { file: file };

    logger
  }

  pub fn log(&mut self, log: String) {
    // no clue why this works
    writeln!(self.file, "{}", log).expect("kaboom");
  }
}
