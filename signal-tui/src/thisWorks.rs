//use color_eyre::Result;
use std::{io, vec};

use crossterm::event::{self, Event, KeyCode, KeyEvent, KeyEventKind};
use ratatui::{
  DefaultTerminal, Frame,
  buffer::Buffer,
  layout::Rect,
  style::Stylize,
  symbols::border,
  text::{Line, Text},
  widgets::{Block, Paragraph, Widget},
};

#[derive(Debug, Default)]
pub struct App {
  counter: u8,
  exit: bool,
  content: Chat,
}

#[derive(Debug, Default, Clone)]
pub struct Message {
  content: String,
  sender: String,
}

#[derive(Debug, Default)]
pub struct Chat {
  id: u8,
  data: Vec<Message>,
}

impl App {
  /// runs the application's main loop until the user quits
  pub fn run(&mut self, terminal: &mut DefaultTerminal) -> io::Result<()> {
    while !self.exit {
      terminal.draw(|frame| self.draw(frame))?;
      self.handle_events()?;
    }
    Ok(())
  }

  fn draw(&self, frame: &mut Frame) {
    frame.render_widget(self, frame.area());
  }

  fn handle_events(&mut self) -> io::Result<()> {
    match event::read()? {
      // it's important to check that the event is a key press event as
      // crossterm also emits key release and repeat events on Windows.
      Event::Key(key_event) if key_event.kind == KeyEventKind::Press => {
        self.handle_key_event(key_event)
      }
      _ => {}
    };
    Ok(())
  }

  fn handle_key_event(&mut self, key_event: KeyEvent) {
    match key_event.code {
      KeyCode::Char('q') => self.exit(),
      KeyCode::Left => self.decrement_counter(),
      KeyCode::Right => self.increment_counter(),
      _ => {}
    }
  }

  fn increment_counter(&mut self) {
    self.counter += 1;
  }

  fn decrement_counter(&mut self) {
    self.counter -= 1;
  }

  fn exit(&mut self) {
    self.exit = true;
  }
}

impl Widget for &App {
  fn render(self, area: Rect, buf: &mut Buffer) {
    let title = Line::from(" Counter App Tutorial ".bold());
    let instructions = Line::from(vec![
      " Decrement ".into(),
      "<Left>".blue().bold(),
      " Increment ".into(),
      "<Right>".blue().bold(),
      " Quit ".into(),
      "<Q> ".blue().bold(),
    ]);
    let block = Block::bordered()
      .title(title.centered())
      .title_bottom(instructions.centered())
      .border_set(border::THICK);

    let _counter_text = Text::from(vec![Line::from(vec![
      "Value: ".into(),
      self.counter.to_string().yellow(),
    ])]);

    //Paragraph::new(counter_text)
    //  .centered()
    //  .block(block)
    //  .render(area, buf);

    let mut messages: Vec<ratatui::prelude::Line> = Vec::new();

    for message in self.content.data.clone() {
      messages.push(Line::from(message.content));
    }

    // let message_text = Text::from(vec![Line::from(vec![
    //   self.content.data[0].content.clone().into(),
    // ])]);
    let message_text = Text::from(messages);

    Paragraph::new(message_text)
      .right_aligned()
      .block(block)
      .render(area, buf);
  }
}

fn main() {
  // let _ = color_eyre::install()
  let mut terminal = ratatui::init();

  let messages = vec![
    Message {
      content: String::from("first message"),
      sender: String::from(""),
    },
    Message {
      content: String::from("second message"),
      sender: String::from(""),
    },
  ];

  let mut chat = Chat::default();

  for message in messages {
    chat.data.push(message);
  }

  let mut app = App::default();
  app.content = chat;

  let _result = app.run(&mut terminal);
  ratatui::restore();
  // result
}
