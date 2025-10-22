#[derive(Debug, Default, Clone)]
pub struct MultiLineString {
  pub body: String,
  cached_lines: Vec<String>,
  cached_width: u16,
  cached_length: u16,
}

impl MultiLineString {
  pub fn init(str: &str) -> Self {
    Self {
      body: str.to_string(),
      cached_lines: vec!["".to_string()],
      cached_width: 0,
      cached_length: 0,
    }
  }

  // this one isnt public cuz smthn smthn object oriented yappery
  fn update_cache(&mut self, width: u16) {
    let mut lines: Vec<String> = Vec::new();
    let mut new_line = String::from("");

    // collumn index
    let mut coldex = 0;
    // let availible_width = (term_width as f32 * settings.message_width_ratio + 0.5) as usize;
    let availible_width = width as usize;

    for yap in self.body.split(" ") {
      let mut length = yap.len();

      if coldex + yap.len() <= availible_width {
        new_line.push_str(yap);
        new_line.push_str(" ");
        coldex += yap.len() + 1;
      } else {
        // INCOMPLETE LOGIC!!! should probably trim the start of the string
        if new_line != "" {
          lines.push(new_line.clone().trim_end().to_string());
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

    lines.push(new_line.clone().trim_end().to_string());

    self.cached_length = self.body.len() as u16;
    self.cached_width = width;
    self.cached_lines = lines;
  }

  // this is the one you call
  pub fn as_lines(&mut self, width: u16) -> &Vec<String> {
    // criteria for refreshing the cache
    if width != self.cached_width || self.body.len() as u16 != self.cached_length {
      self.update_cache(width);
    }

    return &self.cached_lines;
  }
}
