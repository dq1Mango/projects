use std::cmp::min;

use crate::MyStringUtils;

#[derive(Debug, Default, Clone)]
pub struct MultiLineString {
  pub body: String,
  cached_lines: Vec<String>,
  cached_width: u16,
  cached_length: u16,
}

impl MultiLineString {
  pub fn new(str: &str) -> Self {
    Self {
      body: str.to_string(),
      cached_lines: vec!["".to_string()],
      cached_width: 0,
      cached_length: 0,
    }
  }

  fn calc_lines(&self, width: u16) -> Vec<String> {
    let mut lines: Vec<String> = Vec::new();
    let mut new_line = String::from("");

    // collumn index
    let mut coldex = 0;
    // let availible_width = (term_width as f32 * settings.message_width_ratio + 0.5) as usize;
    let availible_width = width as usize;

    // this .split() is a little sketchy but it works mostly
    for yap in self.body.split(" ") {
      let mut length = yap.len();

      if coldex + yap.len() <= availible_width || yap == "" {
        new_line.push_str(yap);
        new_line.push_str(" ");
        coldex += yap.len() + 1;
      } else {
        // INCOMPLETE LOGIC!!!
        if new_line != "" {
          lines.push(new_line.clone());
        }

        let mut index = 0;

        while length >= availible_width {
          lines.push(yap[index..index + availible_width].to_string());
          length -= availible_width;
          index += availible_width;
        }

        new_line = String::from(yap[index..].to_string());
        coldex = new_line.len();

        if new_line.len() > 0 {
          new_line.push_str(" ");
          coldex += 1;
        }
      }
    }

    // remove the trailing ' '
    new_line.pop();
    lines.push(new_line);
    lines
  }

  // this one isnt public cuz smthn smthn object oriented yappery
  fn update_cache(&mut self, width: u16) {
    self.cached_lines = self.calc_lines(width);
    self.cached_length = self.body.len() as u16;
    self.cached_width = width;
  }

  // this is the one you call
  pub fn as_lines(&mut self, width: u16) -> &Vec<String> {
    // criteria for refreshing the cache
    if width != self.cached_width || self.body.len() as u16 != self.cached_length {
      self.update_cache(width);
    }

    return &self.cached_lines;
  }

  pub fn _as_owned_lines(&mut self, width: u16) -> Vec<String> {
    self.as_lines(width).clone()
  }

  pub fn as_trimmed_lines(&mut self, width: u16) -> Vec<String> {
    let untrimmed = self.as_lines(width);
    trim_vec(untrimmed.to_vec())
  }

  pub fn rows(&mut self, width: u16) -> u16 {
    self.as_lines(width).len() as u16
  }

  pub fn fit(&self, width: u16, height: u16) -> Vec<String> {
    let mut fitted = trim_vec(self.calc_lines(width));
    let length = fitted.len();
    fitted = fitted[0..min(height as usize, length)].to_vec();
    // while fitted.len() as u16 > height {
    //   fitted.pop();
    // }

    // let shrunk = fitted[fitted.len() - 1].shrink(width);
    let last = fitted.len() - 1;
    fitted[last] = fitted[last].shrink(width);
    fitted
  }
}

fn trim_vec(untrimmed: Vec<String>) -> Vec<String> {
  let mut trimmed: Vec<String> = vec![];
  for line in untrimmed {
    trimmed.push(line.trim_end().to_string());
  }
  trimmed
}
